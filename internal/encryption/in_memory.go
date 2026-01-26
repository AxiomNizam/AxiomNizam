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

// InMemorySecretsManager in-memory secrets/encryption implementation
type InMemorySecretsManager struct {
	mu        sync.RWMutex
	keys      map[string]*EncryptionKey
	policies  map[string]*EncryptionPolicy
	rotations map[string]*KeyRotation
	auditLogs []*EncryptionAuditLog
}

// NewInMemorySecretsManager creates manager
func NewInMemorySecretsManager() *InMemorySecretsManager {
	return &InMemorySecretsManager{
		keys:      make(map[string]*EncryptionKey),
		policies:  make(map[string]*EncryptionPolicy),
		rotations: make(map[string]*KeyRotation),
		auditLogs: make([]*EncryptionAuditLog, 0),
	}
}

// CreateKey creates encryption key
func (m *InMemorySecretsManager) CreateKey(key *EncryptionKey) (*EncryptionKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key.ID == "" {
		key.ID = fmt.Sprintf("key-%d", time.Now().UnixNano())
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	// Generate actual key material
	keyMaterial := make([]byte, 32) // 256-bit key
	_, err := rand.Read(keyMaterial)
	if err != nil {
		return nil, err
	}
	key.KeyMaterial = base64.StdEncoding.EncodeToString(keyMaterial)

	m.keys[key.ID] = key

	// Log audit
	m.auditLogs = append(m.auditLogs, &EncryptionAuditLog{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		KeyID:     key.ID,
		Action:    "key_created",
		Timestamp: time.Now(),
	})

	return key, nil
}

// GetKey retrieves key
func (m *InMemorySecretsManager) GetKey(id string) (*EncryptionKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[id]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}
	return key, nil
}

// ListKeys lists keys
func (m *InMemorySecretsManager) ListKeys(tenantID string) ([]*EncryptionKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EncryptionKey
	for _, k := range m.keys {
		if tenantID != "" && k.TenantID != tenantID {
			continue
		}
		result = append(result, k)
	}
	return result, nil
}

// RotateKey rotates key
func (m *InMemorySecretsManager) RotateKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, exists := m.keys[id]
	if !exists {
		return fmt.Errorf("key not found")
	}

	// Generate new key material
	newKeyMaterial := make([]byte, 32)
	_, err := rand.Read(newKeyMaterial)
	if err != nil {
		return err
	}

	// Record rotation
	rotation := &KeyRotation{
		ID:        fmt.Sprintf("rotation-%d", time.Now().UnixNano()),
		KeyID:     id,
		OldKeyID:  fmt.Sprintf("%s-v%d", id, key.Version),
		NewKeyID:  fmt.Sprintf("%s-v%d", id, key.Version+1),
		RotatedAt: time.Now(),
	}

	m.rotations[rotation.ID] = rotation

	// Update key
	key.Version++
	key.KeyMaterial = base64.StdEncoding.EncodeToString(newKeyMaterial)
	key.UpdatedAt = time.Now()

	// Log audit
	m.auditLogs = append(m.auditLogs, &EncryptionAuditLog{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		KeyID:     id,
		Action:    "key_rotated",
		Timestamp: time.Now(),
	})

	return nil
}

// Encrypt encrypts data
func (m *InMemorySecretsManager) Encrypt(keyID string, plaintext string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[keyID]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	// Decode key material
	keyMaterial, err := base64.StdEncoding.DecodeString(key.KeyMaterial)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(keyMaterial)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Log audit
	m.auditLogs = append(m.auditLogs, &EncryptionAuditLog{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		KeyID:     keyID,
		Action:    "data_encrypted",
		Timestamp: time.Now(),
	})

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data
func (m *InMemorySecretsManager) Decrypt(keyID string, ciphertext string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[keyID]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	// Decode key material
	keyMaterial, err := base64.StdEncoding.DecodeString(key.KeyMaterial)
	if err != nil {
		return "", err
	}

	// Decode ciphertext
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(keyMaterial)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, ct[:gcm.NonceSize()], ct[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}

	// Log audit
	m.auditLogs = append(m.auditLogs, &EncryptionAuditLog{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		KeyID:     keyID,
		Action:    "data_decrypted",
		Timestamp: time.Now(),
	})

	return string(plaintext), nil
}

// CreatePolicy creates encryption policy
func (m *InMemorySecretsManager) CreatePolicy(policy *EncryptionPolicy) (*EncryptionPolicy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy.ID == "" {
		policy.ID = fmt.Sprintf("policy-%d", time.Now().UnixNano())
	}

	m.policies[policy.ID] = policy
	return policy, nil
}

// GetPolicy retrieves policy
func (m *InMemorySecretsManager) GetPolicy(id string) (*EncryptionPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	policy, exists := m.policies[id]
	if !exists {
		return nil, fmt.Errorf("policy not found")
	}
	return policy, nil
}

// ListPolicies lists policies
func (m *InMemorySecretsManager) ListPolicies(tenantID string) ([]*EncryptionPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EncryptionPolicy
	for _, p := range m.policies {
		if tenantID != "" && p.TenantID != tenantID {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

// DeletePolicy deletes policy
func (m *InMemorySecretsManager) DeletePolicy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.policies, id)
	return nil
}

// GetAuditLogs retrieves audit logs
func (m *InMemorySecretsManager) GetAuditLogs(keyID string) ([]*EncryptionAuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*EncryptionAuditLog
	for _, log := range m.auditLogs {
		if keyID != "" && log.KeyID != keyID {
			continue
		}
		result = append(result, log)
	}
	return result, nil
}
