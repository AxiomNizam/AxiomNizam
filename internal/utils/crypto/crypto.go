package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// PasswordHasher handles password hashing with bcrypt
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// WithCost sets custom bcrypt cost
func (ph *PasswordHasher) WithCost(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	ph.cost = cost
	return ph
}

// Hash hashes a password
func (ph *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// Verify verifies a password against hash
func (ph *PasswordHasher) Verify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Match checks if password matches hash
func (ph *PasswordHasher) Match(password, hash string) bool {
	return ph.Verify(password, hash) == nil
}

// AESEncryption handles AES encryption/decryption
type AESEncryption struct {
	key []byte
}

// NewAESEncryption creates AES encryption with key
func NewAESEncryption(key []byte) (*AESEncryption, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d (must be 16, 24, or 32)", len(key))
	}
	return &AESEncryption{key: key}, nil
}

// Encrypt encrypts data using AES-GCM
func (ae *AESEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(ae.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM
func (ae *AESEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(ae.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string
func (ae *AESEncryption) EncryptString(plaintext string) (string, error) {
	encrypted, err := ae.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a string
func (ae *AESEncryption) DecryptString(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	plaintext, err := ae.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// TokenGenerator generates secure tokens
type TokenGenerator struct {
	size int
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{
		size: 32, // 256 bits
	}
}

// WithSize sets token size in bytes
func (tg *TokenGenerator) WithSize(size int) *TokenGenerator {
	tg.size = size
	return tg
}

// Generate generates a random token
func (tg *TokenGenerator) Generate() (string, error) {
	token := make([]byte, tg.size)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

// MustGenerate generates a token or panics
func (tg *TokenGenerator) MustGenerate() string {
	token, err := tg.Generate()
	if err != nil {
		panic(err)
	}
	return token
}

// Hash represents hash operations
type Hash struct {
	algorithm string
}

// NewHash creates a new hash
func NewHash(algorithm string) *Hash {
	return &Hash{
		algorithm: algorithm,
	}
}

// Compute computes hash
func (h *Hash) Compute(data []byte) []byte {
	switch h.algorithm {
	case "sha256":
		hash := sha256.Sum256(data)
		return hash[:]
	case "sha512":
		hash := sha512.Sum512(data)
		return hash[:]
	default:
		hash := sha256.Sum256(data)
		return hash[:]
	}
}

// ComputeString computes hash of string
func (h *Hash) ComputeString(str string) string {
	hash := h.Compute([]byte(str))
	return base64.StdEncoding.EncodeToString(hash)
}

// PBKDF2 represents PBKDF2 key derivation
type PBKDF2 struct {
	iterations int
	saltSize   int
}

// NewPBKDF2 creates a new PBKDF2
func NewPBKDF2() *PBKDF2 {
	return &PBKDF2{
		iterations: 100000,
		saltSize:   16,
	}
}

// Derive derives a key from password
func (p *PBKDF2) Derive(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, p.iterations, 32, sha256.New)
}

// GenerateSalt generates a random salt
func (p *PBKDF2) GenerateSalt() ([]byte, error) {
	salt := make([]byte, p.saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// RSAKeyPair represents RSA key pair
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// NewRSAKeyPair generates a new RSA key pair
func NewRSAKeyPair(bits int) (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// Encrypt encrypts data with public key
func (rkp *RSAKeyPair) Encrypt(plaintext []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, rkp.PublicKey, plaintext, nil)
}

// Decrypt decrypts data with private key
func (rkp *RSAKeyPair) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, rkp.PrivateKey, ciphertext, nil)
}

// RandomBytes generates random bytes
func RandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

// RandomString generates a random string
func RandomString(size int) (string, error) {
	bytes, err := RandomBytes(size)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SecureCompare compares two byte slices securely
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i] ^ b[i])
	}

	return result == 0
}

// SecureCompareString compares two strings securely
func SecureCompareString(a, b string) bool {
	return SecureCompare([]byte(a), []byte(b))
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Algorithm          string // "aes256", "aes192", "aes128"
	KeySize            int
	SaltSize           int
	Iterations         int
	RequireTLS         bool
	CertificatePinning bool
}

// DefaultEncryptionConfig returns default encryption config
func DefaultEncryptionConfig() EncryptionConfig {
	return EncryptionConfig{
		Algorithm:          "aes256",
		KeySize:            32,
		SaltSize:           16,
		Iterations:         100000,
		RequireTLS:         true,
		CertificatePinning: false,
	}
}
