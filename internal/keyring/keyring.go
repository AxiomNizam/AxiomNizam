// Package keyring implements a rotating-key manager for
// at-rest encryption using AES-GCM.  A key ring has exactly one
// *active* key used for encryption and zero or more *retired* keys
// used only for decrypting older ciphertexts.  Rotation produces a
// new active key; the previously active one becomes retired.
//
// Ciphertexts are prefixed with the 16-byte ID of the key that
// produced them, followed by the 12-byte nonce, followed by the
// AES-GCM output.  Decrypt reads the prefix to find the right key.
//
// The package is modelled on Nomad's keyring subsystem but is
// simplified: it does not persist state (callers are expected to
// serialise the ring via Export / Import) and it does not gossip
// keys between peers.
package keyring

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
)

// KeyID is a 16-byte random identifier for one key version.
type KeyID [16]byte

// String returns the hex encoding — useful for logs and metrics.
func (k KeyID) String() string { return hex.EncodeToString(k[:]) }

// key holds the AEAD and its identifier.
type key struct {
	id   KeyID
	aead cipher.AEAD
}

// Keyring manages active and retired keys.
type Keyring struct {
	mu      sync.RWMutex
	active  *key
	retired map[KeyID]*key
}

// New creates an empty keyring.  Callers must call Rotate once to
// install an initial active key before encrypting.
func New() *Keyring {
	return &Keyring{retired: map[KeyID]*key{}}
}

// Rotate generates a fresh 256-bit key, installs it as active, and
// demotes the current active key to retired.  Returns the new key's
// ID for logging.
func (k *Keyring) Rotate() (KeyID, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return KeyID{}, err
	}
	var id KeyID
	if _, err := io.ReadFull(rand.Reader, id[:]); err != nil {
		return KeyID{}, err
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return KeyID{}, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return KeyID{}, err
	}
	newKey := &key{id: id, aead: aead}

	k.mu.Lock()
	defer k.mu.Unlock()
	if k.active != nil {
		k.retired[k.active.id] = k.active
	}
	k.active = newKey
	return id, nil
}

// ErrNoActiveKey is returned by Encrypt when Rotate has never been called.
var ErrNoActiveKey = errors.New("keyring: no active key — call Rotate first")

// ErrUnknownKey is returned by Decrypt when the ciphertext's key ID
// is not in the ring.  Recovery requires importing the missing key.
var ErrUnknownKey = errors.New("keyring: ciphertext references unknown key")

// Encrypt seals plaintext under the active key and returns
//
//	keyID (16B) || nonce (12B) || AES-GCM(plaintext, aad).
//
// aad may be nil; it is authenticated but not encrypted.
func (k *Keyring) Encrypt(plaintext, aad []byte) ([]byte, error) {
	k.mu.RLock()
	active := k.active
	k.mu.RUnlock()
	if active == nil {
		return nil, ErrNoActiveKey
	}
	nonce := make([]byte, active.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := active.aead.Seal(nil, nonce, plaintext, aad)
	out := make([]byte, 0, len(active.id)+len(nonce)+len(sealed))
	out = append(out, active.id[:]...)
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// Decrypt opens a ciphertext produced by Encrypt.
func (k *Keyring) Decrypt(ciphertext, aad []byte) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, fmt.Errorf("keyring: ciphertext too short (%d bytes)", len(ciphertext))
	}
	var id KeyID
	copy(id[:], ciphertext[:16])
	rest := ciphertext[16:]

	k.mu.RLock()
	var ae cipher.AEAD
	if k.active != nil && k.active.id == id {
		ae = k.active.aead
	} else if r, ok := k.retired[id]; ok {
		ae = r.aead
	}
	k.mu.RUnlock()
	if ae == nil {
		return nil, ErrUnknownKey
	}
	ns := ae.NonceSize()
	if len(rest) < ns {
		return nil, fmt.Errorf("keyring: ciphertext missing nonce")
	}
	nonce := rest[:ns]
	sealed := rest[ns:]
	return ae.Open(nil, nonce, sealed, aad)
}

// ActiveID returns the current active key's ID, or a zero KeyID if none.
func (k *Keyring) ActiveID() KeyID {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if k.active == nil {
		return KeyID{}
	}
	return k.active.id
}

// RetireKey explicitly drops a retired key, rendering any
// ciphertext under that key undecryptable.  Used after a
// re-encrypt-all pass as the final step of key rotation.
func (k *Keyring) RetireKey(id KeyID) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.retired, id)
}
