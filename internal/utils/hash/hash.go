package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
)

// Algorithm defines the hashing algorithm
type Algorithm string

const (
	MD5    Algorithm = "md5"
	SHA1   Algorithm = "sha1"
	SHA256 Algorithm = "sha256"
	SHA512 Algorithm = "sha512"
)

// Hash represents a hash computation
type Hash struct {
	algorithm Algorithm
}

// New creates a new hash computer
func New(algorithm Algorithm) *Hash {
	return &Hash{
		algorithm: algorithm,
	}
}

// Compute computes the hash of data
func (h *Hash) Compute(data []byte) string {
	switch h.algorithm {
	case MD5:
		return fmt.Sprintf("%x", md5.Sum(data))
	case SHA1:
		return fmt.Sprintf("%x", sha1.Sum(data))
	case SHA256:
		return fmt.Sprintf("%x", sha256.Sum256(data))
	case SHA512:
		return fmt.Sprintf("%x", sha512.Sum512(data))
	default:
		return fmt.Sprintf("%x", sha256.Sum256(data))
	}
}

// ComputeString computes the hash of a string
func (h *Hash) ComputeString(str string) string {
	return h.Compute([]byte(str))
}

// ComputeReader computes the hash of data from a reader
func (h *Hash) ComputeReader(reader io.Reader) (string, error) {
	switch h.algorithm {
	case MD5:
		hash := md5.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	case SHA1:
		hash := sha1.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	case SHA256:
		hash := sha256.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	case SHA512:
		hash := sha512.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	default:
		hash := sha256.New()
		if _, err := io.Copy(hash, reader); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	}
}

// HMAC represents HMAC computation
type HMAC struct {
	key       []byte
	algorithm Algorithm
}

// NewHMAC creates a new HMAC computer
func NewHMAC(key []byte, algorithm Algorithm) *HMAC {
	return &HMAC{
		key:       key,
		algorithm: algorithm,
	}
}

// Compute computes the HMAC
func (hm *HMAC) Compute(data []byte) string {
	switch hm.algorithm {
	case MD5:
		hash := hmac.New(md5.New, hm.key)
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case SHA1:
		hash := hmac.New(sha1.New, hm.key)
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case SHA256:
		hash := hmac.New(sha256.New, hm.key)
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	case SHA512:
		hash := hmac.New(sha512.New, hm.key)
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	default:
		hash := hmac.New(sha256.New, hm.key)
		hash.Write(data)
		return hex.EncodeToString(hash.Sum(nil))
	}
}

// ComputeString computes the HMAC of a string
func (hm *HMAC) ComputeString(str string) string {
	return hm.Compute([]byte(str))
}

// Verify verifies the HMAC
func (hm *HMAC) Verify(data []byte, signature string) bool {
	computed := hm.Compute(data)
	return hmac.Equal([]byte(computed), []byte(signature))
}

// VerifyString verifies the HMAC of a string
func (hm *HMAC) VerifyString(str, signature string) bool {
	return hm.Verify([]byte(str), signature)
}

// ConfigFingerprint generates a fingerprint of config using SHA256
func ConfigFingerprint(data []byte) string {
	hash := New(SHA256)
	return hash.Compute(data)
}

// QuickHash quickly computes SHA256 hash of data
func QuickHash(data []byte) string {
	hash := New(SHA256)
	return hash.Compute(data)
}

// QuickHashString quickly computes SHA256 hash of string
func QuickHashString(str string) string {
	hash := New(SHA256)
	return hash.ComputeString(str)
}

// QuickHMAC quickly computes HMAC-SHA256
func QuickHMAC(key, data []byte) string {
	hm := NewHMAC(key, SHA256)
	return hm.Compute(data)
}

// QuickHMACString quickly computes HMAC-SHA256 of string
func QuickHMACString(key string, data string) string {
	return QuickHMAC([]byte(key), []byte(data))
}

// Fingerprint generates a fingerprint of any data
type Fingerprint struct {
	algorithm Algorithm
}

// NewFingerprint creates a new fingerprint generator
func NewFingerprint(algorithm Algorithm) *Fingerprint {
	return &Fingerprint{
		algorithm: algorithm,
	}
}

// Generate generates a fingerprint
func (f *Fingerprint) Generate(data []byte) string {
	hash := New(f.algorithm)
	return hash.Compute(data)
}

// GenerateString generates a fingerprint from string
func (f *Fingerprint) GenerateString(str string) string {
	hash := New(f.algorithm)
	return hash.ComputeString(str)
}

// IsConsistent checks if two fingerprints are consistent
func (f *Fingerprint) IsConsistent(data []byte, expectedFingerprint string) bool {
	computed := f.Generate(data)
	return computed == expectedFingerprint
}

// IsConsistentString checks if two string fingerprints are consistent
func (f *Fingerprint) IsConsistentString(str, expectedFingerprint string) bool {
	computed := f.GenerateString(str)
	return computed == expectedFingerprint
}

// DataIntegrity verifies data integrity using hash
type DataIntegrity struct {
	algorithm Algorithm
}

// NewDataIntegrity creates a new data integrity checker
func NewDataIntegrity(algorithm Algorithm) *DataIntegrity {
	return &DataIntegrity{
		algorithm: algorithm,
	}
}

// ComputeChecksum computes data checksum
func (di *DataIntegrity) ComputeChecksum(data []byte) string {
	hash := New(di.algorithm)
	return hash.Compute(data)
}

// VerifyChecksum verifies data checksum
func (di *DataIntegrity) VerifyChecksum(data []byte, checksum string) bool {
	computed := di.ComputeChecksum(data)
	return computed == checksum
}

// ChecksumMatch checks if two checksums match
func ChecksumMatch(checksum1, checksum2 string) bool {
	return hmac.Equal([]byte(checksum1), []byte(checksum2))
}
