package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"
)

// EncryptionKey represents an encryption key
type EncryptionKey struct {
	ID        string
	Key       []byte
	Algorithm string
	CreatedAt time.Time
	ExpiresAt *time.Time
	IsActive  bool
	Version   int
}

// FieldEncryptionPolicy defines which fields to encrypt
type FieldEncryptionPolicy struct {
	ID              string
	TableName       string
	FieldName       string
	EncryptionType  string // AES256, searchable, deterministic
	KeyID           string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EncryptedField represents encrypted field data
type EncryptedField struct {
	FieldName      string
	EncryptedValue string
	IV             string
	KeyID          string
	Timestamp      time.Time
}

// FieldLevelEncryption manages field-level encryption
type FieldLevelEncryption struct {
	mu               sync.RWMutex
	keys             map[string]*EncryptionKey
	policies         map[string][]*FieldEncryptionPolicy
	encryptedData    map[string]*EncryptedField
	keyRotationLog   []*KeyRotationEvent
	encryptionMetrics *EncryptionMetrics
	maxKeyVersions   int
	maxLogSize       int
}

// KeyRotationEvent logs key rotation
type KeyRotationEvent struct {
	ID            string
	Timestamp     time.Time
	OldKeyID      string
	NewKeyID      string
	FieldsAffected int
	Status        string // started, completed, failed
	ErrorMessage  string
}

// EncryptionMetrics tracks encryption statistics
type EncryptionMetrics struct {
	FieldsEncrypted   int64
	FieldsDecrypted   int64
	KeyRotations      int64
	EncryptionErrors  int64
	DecryptionErrors  int64
	LastKeyRotation   time.Time
	AverageLatency    float64
}

// NewFieldLevelEncryption creates encryption manager
func NewFieldLevelEncryption() *FieldLevelEncryption {
	return &FieldLevelEncryption{
		keys:              make(map[string]*EncryptionKey),
		policies:          make(map[string][]*FieldEncryptionPolicy),
		encryptedData:     make(map[string]*EncryptedField),
		keyRotationLog:    make([]*KeyRotationEvent, 0),
		encryptionMetrics: &EncryptionMetrics{},
		maxKeyVersions:    10,
		maxLogSize:        10000,
	}
}

// RegisterKey registers an encryption key
func (fle *FieldLevelEncryption) RegisterKey(key *EncryptionKey) error {
	fle.mu.Lock()
	defer fle.mu.Unlock()

	if key.ID == "" {
		key.ID = fmt.Sprintf("key-%d", time.Now().UnixNano())
	}

	key.CreatedAt = time.Now()
	fle.keys[key.ID] = key
	return nil
}

// AddEncryptionPolicy adds an encryption policy for a field
func (fle *FieldLevelEncryption) AddEncryptionPolicy(policy *FieldEncryptionPolicy) error {
	fle.mu.Lock()
	defer fle.mu.Unlock()

	if policy.ID == "" {
		policy.ID = fmt.Sprintf("pol-%d", time.Now().UnixNano())
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	tableField := fmt.Sprintf("%s.%s", policy.TableName, policy.FieldName)

	if _, exists := fle.policies[tableField]; !exists {
		fle.policies[tableField] = make([]*FieldEncryptionPolicy, 0)
	}

	fle.policies[tableField] = append(fle.policies[tableField], policy)
	return nil
}

// EncryptField encrypts a field value
func (fle *FieldLevelEncryption) EncryptField(tableField string, value interface{}, keyID string) (*EncryptedField, error) {
	fle.mu.RLock()
	key, exists := fle.keys[keyID]
	if !exists {
		fle.mu.RUnlock()
		fle.mu.Lock()
		fle.encryptionMetrics.EncryptionErrors++
		fle.mu.Unlock()
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	if !key.IsActive {
		fle.mu.RUnlock()
		return nil, fmt.Errorf("key not active: %s", keyID)
	}
	fle.mu.RUnlock()

	// Convert value to string
	plaintext := []byte(fmt.Sprintf("%v", value))

	// Create AES cipher
	block, err := aes.NewCipher(key.Key)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.EncryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.EncryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.EncryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	encryptedValue := base64.StdEncoding.EncodeToString(ciphertext)
	ivValue := base64.StdEncoding.EncodeToString(nonce)

	encryptedField := &EncryptedField{
		FieldName:      tableField,
		EncryptedValue: encryptedValue,
		IV:             ivValue,
		KeyID:          keyID,
		Timestamp:      time.Now(),
	}

	fle.mu.Lock()
	fle.encryptedData[tableField] = encryptedField
	fle.encryptionMetrics.FieldsEncrypted++
	fle.mu.Unlock()

	return encryptedField, nil
}

// DecryptField decrypts a field value
func (fle *FieldLevelEncryption) DecryptField(encryptedField *EncryptedField) (interface{}, error) {
	fle.mu.RLock()
	key, exists := fle.keys[encryptedField.KeyID]
	if !exists {
		fle.mu.RUnlock()
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, fmt.Errorf("key not found: %s", encryptedField.KeyID)
	}
	fle.mu.RUnlock()

	// Decode ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedField.EncryptedValue)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Decode IV
	iv, err := base64.StdEncoding.DecodeString(encryptedField.IV)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(key.Key)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		fle.mu.Lock()
		fle.encryptionMetrics.DecryptionErrors++
		fle.mu.Unlock()
		return nil, err
	}

	fle.mu.Lock()
	fle.encryptionMetrics.FieldsDecrypted++
	fle.mu.Unlock()

	return string(plaintext), nil
}

// RotateKey rotates encryption key
func (fle *FieldLevelEncryption) RotateKey(oldKeyID string, newKey *EncryptionKey) (*KeyRotationEvent, error) {
	fle.mu.Lock()
	defer fle.mu.Unlock()

	event := &KeyRotationEvent{
		ID:        fmt.Sprintf("rot-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		OldKeyID:  oldKeyID,
		Status:    "started",
	}

	// Register new key
	if newKey.ID == "" {
		newKey.ID = fmt.Sprintf("key-%d", time.Now().UnixNano())
	}
	newKey.CreatedAt = time.Now()
	fle.keys[newKey.ID] = newKey
	event.NewKeyID = newKey.ID

	// Mark old key inactive
	if oldKey, exists := fle.keys[oldKeyID]; exists {
		oldKey.IsActive = false
	}

	event.Status = "completed"
	fle.keyRotationLog = append(fle.keyRotationLog, event)
	fle.encryptionMetrics.KeyRotations++
	fle.encryptionMetrics.LastKeyRotation = time.Now()

	if len(fle.keyRotationLog) > fle.maxLogSize {
		fle.keyRotationLog = fle.keyRotationLog[1:]
	}

	return event, nil
}

// GetPoliciesForField gets encryption policies for a field
func (fle *FieldLevelEncryption) GetPoliciesForField(table, field string) []*FieldEncryptionPolicy {
	fle.mu.RLock()
	defer fle.mu.RUnlock()

	tableField := fmt.Sprintf("%s.%s", table, field)
	if policies, exists := fle.policies[tableField]; exists {
		return policies
	}
	return make([]*FieldEncryptionPolicy, 0)
}

// GetEncryptionMetrics returns encryption metrics
func (fle *FieldLevelEncryption) GetEncryptionMetrics() *EncryptionMetrics {
	fle.mu.RLock()
	defer fle.mu.RUnlock()

	return &EncryptionMetrics{
		FieldsEncrypted:  fle.encryptionMetrics.FieldsEncrypted,
		FieldsDecrypted:  fle.encryptionMetrics.FieldsDecrypted,
		KeyRotations:     fle.encryptionMetrics.KeyRotations,
		EncryptionErrors: fle.encryptionMetrics.EncryptionErrors,
		DecryptionErrors: fle.encryptionMetrics.DecryptionErrors,
		LastKeyRotation:  fle.encryptionMetrics.LastKeyRotation,
	}
}

// GetKeyRotationLog gets key rotation history
func (fle *FieldLevelEncryption) GetKeyRotationLog(limit int) []*KeyRotationEvent {
	fle.mu.RLock()
	defer fle.mu.RUnlock()

	if limit > len(fle.keyRotationLog) {
		limit = len(fle.keyRotationLog)
	}
	if limit == 0 {
		return make([]*KeyRotationEvent, 0)
	}

	return fle.keyRotationLog[len(fle.keyRotationLog)-limit:]
}

// ListActiveKeys lists active encryption keys
func (fle *FieldLevelEncryption) ListActiveKeys() []*EncryptionKey {
	fle.mu.RLock()
	defer fle.mu.RUnlock()

	activeKeys := make([]*EncryptionKey, 0)
	for _, key := range fle.keys {
		if key.IsActive {
			activeKeys = append(activeKeys, key)
		}
	}
	return activeKeys
}

// GetEncryptionStatus returns overall encryption status
func (fle *FieldLevelEncryption) GetEncryptionStatus() map[string]interface{} {
	fle.mu.RLock()
	defer fle.mu.RUnlock()

	successRate := 0.0
	totalOps := fle.encryptionMetrics.FieldsEncrypted + fle.encryptionMetrics.FieldsDecrypted
	if totalOps > 0 {
		successRate = float64(totalOps) / float64(totalOps+fle.encryptionMetrics.EncryptionErrors+fle.encryptionMetrics.DecryptionErrors) * 100
	}

	return map[string]interface{}{
		"active_keys":          len(fle.ListActiveKeys()),
		"total_policies":       len(fle.policies),
		"fields_encrypted":     fle.encryptionMetrics.FieldsEncrypted,
		"fields_decrypted":     fle.encryptionMetrics.FieldsDecrypted,
		"key_rotations":        fle.encryptionMetrics.KeyRotations,
		"encryption_errors":    fle.encryptionMetrics.EncryptionErrors,
		"decryption_errors":    fle.encryptionMetrics.DecryptionErrors,
		"success_rate":         successRate,
		"last_key_rotation":    fle.encryptionMetrics.LastKeyRotation,
	}
}
