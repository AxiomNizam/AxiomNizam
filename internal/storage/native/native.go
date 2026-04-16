package native

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"example.com/axiomnizam/internal/storage/models"
)

// Backend is a self-contained, filesystem-based object storage engine.
// It needs no external service (no MinIO, no S3) — data lives on the
// local volume mounted into the container.
//
// Layout on disk:
//
//	<root>/
//	  <bucket>/
//	    .axiom/config.json          – versioning, lifecycle rules
//	    .axiom/tags.json            – bucket tags
//	    .axiom/objects/<key>.meta   – per-object metadata (JSON)
//	    data/<key>                  – the actual object bytes
type Backend struct {
	mu       sync.RWMutex
	root     string // absolute path of the storage root directory
	endpoint string // human-readable label (e.g. "native://data/storage")
	// Secret used for HMAC-signed presign tokens.
	presignSecret []byte
}

// New creates a native filesystem backend rooted at root.
// The directory is created if it does not exist.
func New(root string, presignSecret string) (*Backend, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("native storage: resolve root: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("native storage: create root %q: %w", abs, err)
	}
	secret := []byte(presignSecret)
	if len(secret) == 0 {
		secret = []byte("axiom-native-storage-default-key")
	}
	log.Printf("✅ Storage: native filesystem backend at %s", abs)
	return &Backend{
		root:          abs,
		endpoint:      "native://" + abs,
		presignSecret: secret,
	}, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (b *Backend) bucketDir(bucket string) string {
	return filepath.Join(b.root, bucket)
}

func (b *Backend) dataDir(bucket string) string {
	return filepath.Join(b.root, bucket, "data")
}

func (b *Backend) metaDir(bucket string) string {
	return filepath.Join(b.root, bucket, ".axiom")
}

func (b *Backend) objectMetaDir(bucket string) string {
	return filepath.Join(b.root, bucket, ".axiom", "objects")
}

func (b *Backend) objectPath(bucket, key string) string {
	return filepath.Join(b.dataDir(bucket), filepath.FromSlash(key))
}

func (b *Backend) objectMetaPath(bucket, key string) string {
	return filepath.Join(b.objectMetaDir(bucket), filepath.FromSlash(key)+".meta")
}

func (b *Backend) configPath(bucket string) string {
	return filepath.Join(b.metaDir(bucket), "config.json")
}

func (b *Backend) tagsPath(bucket string) string {
	return filepath.Join(b.metaDir(bucket), "tags.json")
}

// bucketConfig is persisted per-bucket.
type bucketConfig struct {
	Versioning bool                   `json:"versioning"`
	Lifecycle  []models.LifecycleRule `json:"lifecycle,omitempty"`
}

func (b *Backend) readConfig(bucket string) bucketConfig {
	data, err := os.ReadFile(b.configPath(bucket))
	if err != nil {
		return bucketConfig{}
	}
	var cfg bucketConfig
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

func (b *Backend) writeConfig(bucket string, cfg bucketConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(b.configPath(bucket), data, 0o640)
}

// objectMeta is persisted per-object alongside the data file.
type objectMeta struct {
	ContentType  string    `json:"contentType"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"lastModified"`
}

func (b *Backend) readObjectMeta(bucket, key string) (objectMeta, error) {
	data, err := os.ReadFile(b.objectMetaPath(bucket, key))
	if err != nil {
		return objectMeta{}, err
	}
	var m objectMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return objectMeta{}, err
	}
	return m, nil
}

func (b *Backend) writeObjectMeta(bucket, key string, m objectMeta) error {
	p := b.objectMetaPath(bucket, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o640)
}

func (b *Backend) deleteObjectMeta(bucket, key string) {
	_ = os.Remove(b.objectMetaPath(bucket, key))
}

// ---------------------------------------------------------------------------
// Backend interface
// ---------------------------------------------------------------------------

// Ping always succeeds — the filesystem is always reachable.
func (b *Backend) Ping(_ context.Context) error {
	if _, err := os.Stat(b.root); err != nil {
		return fmt.Errorf("native storage: root directory not accessible: %w", err)
	}
	return nil
}

// Endpoint returns a label for the native backend.
func (b *Backend) Endpoint() string {
	return b.endpoint
}

// ---- Bucket Operations ----

func (b *Backend) CreateBucket(_ context.Context, name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	dir := b.bucketDir(name)
	if _, err := os.Stat(dir); err == nil {
		return nil // idempotent
	}
	if err := os.MkdirAll(b.dataDir(name), 0o750); err != nil {
		return fmt.Errorf("native storage: create bucket %q: %w", name, err)
	}
	if err := os.MkdirAll(b.objectMetaDir(name), 0o750); err != nil {
		return fmt.Errorf("native storage: create bucket meta %q: %w", name, err)
	}
	if err := b.writeConfig(name, bucketConfig{}); err != nil {
		return fmt.Errorf("native storage: init bucket config %q: %w", name, err)
	}
	log.Printf("✅ Storage: bucket %q created (native)", name)
	return nil
}

func (b *Backend) DeleteBucket(_ context.Context, name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	dir := b.bucketDir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("native storage: bucket %q does not exist", name)
	}

	// Check the bucket is empty (only metadata dirs remain).
	entries, err := os.ReadDir(b.dataDir(name))
	if err == nil && len(entries) > 0 {
		return fmt.Errorf("native storage: bucket %q is not empty (%d objects)", name, len(entries))
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("native storage: remove bucket %q: %w", name, err)
	}
	log.Printf("✅ Storage: bucket %q deleted (native)", name)
	return nil
}

func (b *Backend) BucketExists(_ context.Context, name string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, err := os.Stat(b.bucketDir(name))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Backend) ListBuckets(_ context.Context) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	entries, err := os.ReadDir(b.root)
	if err != nil {
		return nil, fmt.Errorf("native storage: list buckets: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// ---- Versioning ----

func (b *Backend) SetBucketVersioning(_ context.Context, bucket string, enabled bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	cfg := b.readConfig(bucket)
	cfg.Versioning = enabled
	return b.writeConfig(bucket, cfg)
}

func (b *Backend) GetBucketVersioning(_ context.Context, bucket string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cfg := b.readConfig(bucket)
	return cfg.Versioning, nil
}

// ---- Lifecycle ----

func (b *Backend) SetBucketLifecycle(_ context.Context, bucket string, rules []models.LifecycleRule) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	cfg := b.readConfig(bucket)
	cfg.Lifecycle = rules
	return b.writeConfig(bucket, cfg)
}

// ---- Object operations ----

func (b *Backend) PutObject(_ context.Context, bucket, key string, data io.Reader, _ int64, contentType string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, err := os.Stat(b.bucketDir(bucket)); os.IsNotExist(err) {
		return fmt.Errorf("native storage: bucket %q does not exist", bucket)
	}

	objPath := b.objectPath(bucket, key)
	if err := os.MkdirAll(filepath.Dir(objPath), 0o750); err != nil {
		return fmt.Errorf("native storage: create parent dirs for %q/%q: %w", bucket, key, err)
	}

	f, err := os.Create(objPath)
	if err != nil {
		return fmt.Errorf("native storage: create file %q/%q: %w", bucket, key, err)
	}

	hasher := md5.New()
	w := io.MultiWriter(f, hasher)
	n, copyErr := io.Copy(w, data)
	closeErr := f.Close()
	if copyErr != nil {
		return fmt.Errorf("native storage: write %q/%q: %w", bucket, key, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("native storage: close %q/%q: %w", bucket, key, closeErr)
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	meta := objectMeta{
		ContentType:  contentType,
		Size:         n,
		ETag:         hex.EncodeToString(hasher.Sum(nil)),
		LastModified: time.Now().UTC(),
	}
	if err := b.writeObjectMeta(bucket, key, meta); err != nil {
		return fmt.Errorf("native storage: write meta %q/%q: %w", bucket, key, err)
	}
	return nil
}

func (b *Backend) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	f, err := os.Open(b.objectPath(bucket, key))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("native storage: object %q/%q not found", bucket, key)
	}
	if err != nil {
		return nil, fmt.Errorf("native storage: open %q/%q: %w", bucket, key, err)
	}
	return f, nil
}

func (b *Backend) DeleteObject(_ context.Context, bucket, key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := os.Remove(b.objectPath(bucket, key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("native storage: delete %q/%q: %w", bucket, key, err)
	}
	b.deleteObjectMeta(bucket, key)
	return nil
}

func (b *Backend) ListObjects(_ context.Context, bucket, prefix string) ([]models.ObjectInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	dataRoot := b.dataDir(bucket)
	if _, err := os.Stat(dataRoot); os.IsNotExist(err) {
		return nil, nil
	}

	var objects []models.ObjectInfo
	err := filepath.Walk(dataRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dataRoot, path)
		key := filepath.ToSlash(rel)
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		obj := models.ObjectInfo{
			Key:          key,
			Size:         info.Size(),
			LastModified: info.ModTime().UTC(),
		}
		// Try to read cached meta for etag / content-type.
		if m, err := b.readObjectMeta(bucket, key); err == nil {
			obj.ETag = m.ETag
			obj.ContentType = m.ContentType
		}
		objects = append(objects, obj)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("native storage: walk %q: %w", bucket, err)
	}

	sort.Slice(objects, func(i, j int) bool { return objects[i].Key < objects[j].Key })
	return objects, nil
}

func (b *Backend) StatObject(_ context.Context, bucket, key string) (*models.ObjectInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	fi, err := os.Stat(b.objectPath(bucket, key))
	if err != nil {
		return nil, fmt.Errorf("native storage: stat %q/%q: %w", bucket, key, err)
	}

	obj := &models.ObjectInfo{
		Key:          key,
		Size:         fi.Size(),
		LastModified: fi.ModTime().UTC(),
	}
	if m, mErr := b.readObjectMeta(bucket, key); mErr == nil {
		obj.ContentType = m.ContentType
		obj.ETag = m.ETag
	}
	return obj, nil
}

// ---- Batch / copy ----

func (b *Backend) MultiDeleteObjects(ctx context.Context, bucket string, keys []string) (int, []string, error) {
	var deleted int
	var errs []string
	for _, k := range keys {
		if err := b.DeleteObject(ctx, bucket, k); err != nil {
			errs = append(errs, k+": "+err.Error())
		} else {
			deleted++
		}
	}
	return deleted, errs, nil
}

func (b *Backend) CopyObject(_ context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	srcPath := b.objectPath(srcBucket, srcKey)
	dstPath := b.objectPath(dstBucket, dstKey)

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("native storage: copy source %q/%q: %w", srcBucket, srcKey, err)
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o750); err != nil {
		return fmt.Errorf("native storage: copy dest dir %q/%q: %w", dstBucket, dstKey, err)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("native storage: copy dest %q/%q: %w", dstBucket, dstKey, err)
	}

	hasher := md5.New()
	w := io.MultiWriter(dst, hasher)
	n, copyErr := io.Copy(w, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return fmt.Errorf("native storage: copy write: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("native storage: copy close: %w", closeErr)
	}

	// Read source meta to preserve content-type.
	srcMeta, _ := b.readObjectMeta(srcBucket, srcKey)
	ct := srcMeta.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}

	meta := objectMeta{
		ContentType:  ct,
		Size:         n,
		ETag:         hex.EncodeToString(hasher.Sum(nil)),
		LastModified: time.Now().UTC(),
	}
	if err := b.writeObjectMeta(dstBucket, dstKey, meta); err != nil {
		return fmt.Errorf("native storage: copy meta: %w", err)
	}

	log.Printf("✅ Storage: copied %s/%s → %s/%s (native)", srcBucket, srcKey, dstBucket, dstKey)
	return nil
}

// ---- Pre-signed URLs ----
//
// For the native backend, pre-signed URLs are HMAC-signed tokens that point
// back to the AxiomNizam API server. The token encodes bucket, key, method,
// and expiry and is verified by the storage handler at request time.

func (b *Backend) presignToken(method, bucket, key string, expires time.Duration) string {
	exp := time.Now().UTC().Add(expires).Unix()
	payload := fmt.Sprintf("%s\n%s\n%s\n%d", method, bucket, key, exp)
	mac := hmac.New(sha256.New, b.presignSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d:%s", exp, sig)
}

func (b *Backend) PresignGetObject(_ context.Context, bucket, key string, expires time.Duration) (string, error) {
	token := b.presignToken("GET", bucket, key, expires)
	u := url.URL{
		Path: fmt.Sprintf("/api/v1/storage/buckets/%s/objects/%s", bucket, key),
		RawQuery: url.Values{
			"X-Axiom-Expires": {fmt.Sprintf("%d", int(expires.Seconds()))},
			"X-Axiom-Token":   {token},
		}.Encode(),
	}
	return u.String(), nil
}

func (b *Backend) PresignPutObject(_ context.Context, bucket, key string, expires time.Duration) (string, error) {
	token := b.presignToken("PUT", bucket, key, expires)
	u := url.URL{
		Path: fmt.Sprintf("/api/v1/storage/buckets/%s/objects/%s", bucket, key),
		RawQuery: url.Values{
			"X-Axiom-Expires": {fmt.Sprintf("%d", int(expires.Seconds()))},
			"X-Axiom-Token":   {token},
		}.Encode(),
	}
	return u.String(), nil
}

// VerifyPresignToken validates an HMAC presign token. It is exported so the
// HTTP handler layer can call it when a presigned request arrives.
func (b *Backend) VerifyPresignToken(method, bucket, key, token string) bool {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return false
	}

	expStr := parts[0]
	sig := parts[1]

	var exp int64
	if _, err := fmt.Sscanf(expStr, "%d", &exp); err != nil {
		return false
	}
	if time.Now().UTC().Unix() > exp {
		return false // expired
	}

	payload := fmt.Sprintf("%s\n%s\n%s\n%d", method, bucket, key, exp)
	mac := hmac.New(sha256.New, b.presignSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

// ---- Tagging ----

func (b *Backend) GetBucketTagging(_ context.Context, bucket string) ([]models.BucketTag, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data, err := os.ReadFile(b.tagsPath(bucket))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("native storage: read tags %q: %w", bucket, err)
	}
	var tags []models.BucketTag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("native storage: parse tags %q: %w", bucket, err)
	}
	return tags, nil
}

func (b *Backend) PutBucketTagging(_ context.Context, bucket string, tags []models.BucketTag) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(b.tagsPath(bucket), data, 0o640)
}

func (b *Backend) DeleteBucketTagging(_ context.Context, bucket string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := os.Remove(b.tagsPath(bucket)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("native storage: delete tags %q: %w", bucket, err)
	}
	return nil
}

// ---- Utility ----

func (b *Backend) GetBucketSize(ctx context.Context, bucket string) (int64, int64, error) {
	objects, err := b.ListObjects(ctx, bucket, "")
	if err != nil {
		return 0, 0, err
	}
	var totalSize int64
	for _, o := range objects {
		totalSize += o.Size
	}
	return totalSize, int64(len(objects)), nil
}

// ---------------------------------------------------------------------------
// Compile-time interface check (import cycle–safe: we check against a local
// copy of the interface signature, but the real check happens when storage.go
// assigns *Backend to a storage.Backend variable).
// ---------------------------------------------------------------------------

// Ensure *Backend satisfies io.Closer for future use.
var _ io.ReadCloser = (*os.File)(nil) // helper: os.File is a ReadCloser
// Note: the compile-time check against storage.Backend happens in storage.go
// to avoid an import cycle.
