package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configurable parameters for the SafeGate scanner pipeline.
// Every field has a sensible default. Override via environment variables.
type Config struct {
	// ── Global settings ──────────────────────────────────────────────────

	// Timeout is the maximum time the orchestrator will wait for all
	// scanners to complete. Individual scanners may finish sooner.
	// Env: SCANNER_TIMEOUT (default: 2m)
	Timeout time.Duration

	// Parallel controls whether scanners run concurrently (true) or
	// sequentially (false). Sequential mode is useful for debugging.
	// Env: SCANNER_PARALLEL (default: true)
	Parallel bool

	// ── Metadata scanner ─────────────────────────────────────────────────

	// MaxFileSize is the maximum allowed file size in bytes.
	// Files exceeding this limit receive a high-severity finding.
	// Env: SCANNER_MAX_FILE_SIZE (default: 104857600 = 100MB)
	MaxFileSize int64

	// NullByteSampleSize is the number of bytes to sample when checking
	// text files for null byte injection. Sampling avoids O(n) on large files.
	// Set to 0 to scan entire file content (legacy behavior).
	// Env: SCANNER_NULL_BYTE_SAMPLE_SIZE (default: 8192)
	NullByteSampleSize int

	// MaxFilenameLength is the maximum allowed filename length.
	// Env: SCANNER_MAX_FILENAME_LENGTH (default: 255)
	MaxFilenameLength int

	// ── MIME scanner ─────────────────────────────────────────────────────

	// AllowedMIMETypes is the list of MIME types permitted for upload.
	// Files whose detected content type is not in this list receive a
	// high-severity finding.
	// Env: SCANNER_ALLOWED_MIME_TYPES (default: built-in list)
	AllowedMIMETypes []string

	// ── Archive scanner ──────────────────────────────────────────────────

	// ArchiveMaxDepth is the maximum nesting depth for recursive archive
	// inspection. Exceeding this depth triggers a high-severity finding.
	// Env: SCANNER_ARCHIVE_MAX_DEPTH (default: 5)
	ArchiveMaxDepth int

	// ArchiveMaxDecompressedSize is the maximum total decompressed size
	// in bytes across all entries in an archive. Exceeding this triggers
	// a critical finding (zip bomb protection).
	// Env: SCANNER_ARCHIVE_MAX_DECOMPRESS (default: 1073741824 = 1GB)
	ArchiveMaxDecompressedSize int64

	// ArchiveMaxFiles is the maximum number of entries allowed in an archive.
	// Exceeding this triggers a high-severity finding.
	// Env: SCANNER_ARCHIVE_MAX_FILES (default: 10000)
	ArchiveMaxFiles int

	// ArchiveCompressionRatioLimit is the maximum compression ratio (uncompressed/compressed)
	// before a file is flagged as a potential zip bomb.
	// Env: SCANNER_ARCHIVE_RATIO_LIMIT (default: 100)
	ArchiveCompressionRatioLimit float64
}

// DefaultConfig returns a Config populated with production-safe defaults.
func DefaultConfig() Config {
	return Config{
		Timeout:            2 * time.Minute,
		Parallel:           true,
		MaxFileSize:        100 * 1024 * 1024, // 100MB
		NullByteSampleSize: 8192,
		MaxFilenameLength:  255,
		AllowedMIMETypes:   defaultAllowedMIMETypes(),
		ArchiveMaxDepth:    5,
		ArchiveMaxDecompressedSize: 1024 * 1024 * 1024, // 1GB
		ArchiveMaxFiles:             10000,
		ArchiveCompressionRatioLimit: 100.0,
	}
}

// LoadFromEnv creates a Config from DefaultConfig() and overrides
// any fields that have corresponding environment variables set.
func LoadFromEnv() Config {
	cfg := DefaultConfig()

	if v := envDuration("SCANNER_TIMEOUT"); v > 0 {
		cfg.Timeout = v
	}
	if v, ok := envBool("SCANNER_PARALLEL"); ok {
		cfg.Parallel = v
	}
	if v := envInt64("SCANNER_MAX_FILE_SIZE"); v > 0 {
		cfg.MaxFileSize = v
	}
	if v := envInt("SCANNER_NULL_BYTE_SAMPLE_SIZE"); v >= 0 {
		cfg.NullByteSampleSize = v
	}
	if v := envInt("SCANNER_MAX_FILENAME_LENGTH"); v > 0 {
		cfg.MaxFilenameLength = v
	}
	if v := envStringSlice("SCANNER_ALLOWED_MIME_TYPES"); len(v) > 0 {
		cfg.AllowedMIMETypes = v
	}
	if v := envInt("SCANNER_ARCHIVE_MAX_DEPTH"); v > 0 {
		cfg.ArchiveMaxDepth = v
	}
	if v := envInt64("SCANNER_ARCHIVE_MAX_DECOMPRESS"); v > 0 {
		cfg.ArchiveMaxDecompressedSize = v
	}
	if v := envInt("SCANNER_ARCHIVE_MAX_FILES"); v > 0 {
		cfg.ArchiveMaxFiles = v
	}
	if v := envFloat64("SCANNER_ARCHIVE_RATIO_LIMIT"); v > 0 {
		cfg.ArchiveCompressionRatioLimit = v
	}

	return cfg
}

// Validate checks the configuration for invalid values.
func (c Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("scanner: timeout must be positive, got %v", c.Timeout)
	}
	if c.MaxFileSize <= 0 {
		return fmt.Errorf("scanner: max file size must be positive, got %d", c.MaxFileSize)
	}
	if c.ArchiveMaxDepth <= 0 {
		return fmt.Errorf("scanner: archive max depth must be positive, got %d", c.ArchiveMaxDepth)
	}
	if c.ArchiveCompressionRatioLimit <= 0 {
		return fmt.Errorf("scanner: archive ratio limit must be positive, got %f", c.ArchiveCompressionRatioLimit)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Default MIME type list
// ─────────────────────────────────────────────────────────────────────────────

func defaultAllowedMIMETypes() []string {
	return []string{
		// Text
		"text/plain", "text/csv", "text/html", "text/xml",
		// Documents
		"application/json", "application/xml", "application/pdf",
		// Archives
		"application/zip", "application/gzip",
		// Office — modern (OOXML)
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		// Office — legacy
		"application/vnd.ms-excel",
		"application/msword",
		"application/vnd.ms-powerpoint",
		// Images
		"image/png", "image/jpeg", "image/gif", "image/svg+xml", "image/webp",
		// Media
		"audio/mpeg", "video/mp4",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Environment variable helpers
// ─────────────────────────────────────────────────────────────────────────────

func envDuration(key string) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}

func envBool(key string) (bool, bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return false, false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return b, true
}

func envInt(key string) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return -1
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return -1
	}
	return n
}

func envInt64(key string) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func envFloat64(key string) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

func envStringSlice(key string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
