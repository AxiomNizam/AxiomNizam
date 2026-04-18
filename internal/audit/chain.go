// Package audit — tamper-evident hash chain.
//
// Each audit entry is canonicalised, serialised, and hashed together
// with the previous entry's hash.  The resulting SHA-256 is stored on
// the entry's ImmutableHash field.  Verification walks the chain in
// order and recomputes each hash; any divergence identifies the
// earliest tampered entry.
//
// Threat model
//
//   - Protects against silent modification of historical entries.  An
//     attacker who rewrites entry N must also rewrite every subsequent
//     hash up to the chain head, and must also rewrite any external
//     attestation of the head hash.  Exporting the head hash to an
//     append-only external system (syslog, S3 object-lock bucket,
//     blockchain timestamping service) closes that loop.
//
//   - Does NOT protect against append-time suppression: an attacker
//     with write access to the underlying store can refuse to persist
//     an entry.  Mitigations for that failure mode live in the sink
//     layer (replicated writes, quorum acks).
//
// # Canonicalisation
//
// The map fields OldValues, NewValues, Changes, and Labels are serialised
// with sorted keys so that re-serialisation yields an identical byte
// stream.  Fields that are volatile or derived (ImmutableHash itself)
// are excluded from the digest.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// GenesisHash is the seed value used as the "previous hash" of the
// first entry in any chain.  Changing this constant invalidates all
// existing chains — do not alter after deployment.
const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// canonicalLog is the stable projection of AuditLog used for hashing.
// Fields are exported so encoding/json can populate them; field names
// are intentionally terse to minimise the hash input size.
type canonicalLog struct {
	ID           string                 `json:"id"`
	T            string                 `json:"t"` // RFC3339Nano timestamp
	Tenant       string                 `json:"tn,omitempty"`
	User         string                 `json:"u,omitempty"`
	Name         string                 `json:"un,omitempty"`
	Action       string                 `json:"a"`
	Result       string                 `json:"r"`
	ResourceType string                 `json:"rt,omitempty"`
	ResourceID   string                 `json:"ri,omitempty"`
	ResourceName string                 `json:"rn,omitempty"`
	Namespace    string                 `json:"ns,omitempty"`
	OldValues    map[string]interface{} `json:"ov,omitempty"`
	NewValues    map[string]interface{} `json:"nv,omitempty"`
	Changes      []Change               `json:"ch,omitempty"`
	SourceIP     string                 `json:"ip,omitempty"`
	Method       string                 `json:"m,omitempty"`
	Path         string                 `json:"p,omitempty"`
	RequestID    string                 `json:"rq,omitempty"`
	StatusCode   int                    `json:"s,omitempty"`
	ErrorMessage string                 `json:"e,omitempty"`
	Reason       string                 `json:"rs,omitempty"`
	Labels       map[string]string      `json:"l,omitempty"`
	Prev         string                 `json:"prev"`
}

func toCanonical(e *AuditLog, prev string) canonicalLog {
	return canonicalLog{
		ID:           e.ID,
		T:            e.Timestamp.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
		Tenant:       e.TenantID,
		User:         e.UserID,
		Name:         e.Username,
		Action:       string(e.Action),
		Result:       string(e.Result),
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		ResourceName: e.ResourceName,
		Namespace:    e.Namespace,
		OldValues:    e.OldValues,
		NewValues:    e.NewValues,
		Changes:      e.Changes,
		SourceIP:     e.SourceIP,
		Method:       e.Method,
		Path:         e.Path,
		RequestID:    e.RequestID,
		StatusCode:   e.StatusCode,
		ErrorMessage: e.ErrorMessage,
		Reason:       e.Reason,
		Labels:       e.Labels,
		Prev:         prev,
	}
}

// marshalSorted emits a JSON encoding whose map keys are in
// lexicographic order.  encoding/json already sorts top-level map
// keys, but nested maps within interface{} values are traversed
// recursively to ensure a fully deterministic byte stream.
func marshalSorted(v canonicalLog) ([]byte, error) {
	// Normalise nested maps first.
	v.OldValues = sortMap(v.OldValues)
	v.NewValues = sortMap(v.NewValues)
	return json.Marshal(v)
}

// sortMap rebuilds m so that any nested map[string]interface{} values
// are traversed in sorted order.  The returned map itself has no
// defined iteration order — ordering is enforced by marshalling via
// encoding/json which sorts keys alphabetically for map types.
func sortMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch v := m[k].(type) {
		case map[string]interface{}:
			out[k] = sortMap(v)
		default:
			out[k] = v
		}
	}
	return out
}

// ComputeHash returns the SHA-256 hex digest for entry given the
// previous entry's hash.  The entry's ImmutableHash field is NOT
// consulted — callers pass its predecessor explicitly.
func ComputeHash(entry *AuditLog, prevHash string) (string, error) {
	if prevHash == "" {
		prevHash = GenesisHash
	}
	canon := toCanonical(entry, prevHash)
	buf, err := marshalSorted(canon)
	if err != nil {
		return "", fmt.Errorf("canonicalise audit entry %q: %w", entry.ID, err)
	}
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:]), nil
}

// Seal computes and stores the chain hash for entry.  Callers invoke
// Seal immediately before persisting the entry to the underlying store;
// the updated ImmutableHash becomes the "previous hash" for the next
// entry in the chain.
func Seal(entry *AuditLog, prevHash string) error {
	h, err := ComputeHash(entry, prevHash)
	if err != nil {
		return err
	}
	entry.ImmutableHash = h
	return nil
}

// VerificationResult describes the outcome of a chain verification
// pass.  On success Broken is the zero value; on failure Broken points
// to the earliest entry whose recomputed hash does not match its
// stored ImmutableHash, and Reason explains the mismatch.
type VerificationResult struct {
	Entries int
	OK      bool
	Broken  *AuditLog
	Reason  string
}

// VerifyChain walks entries in the order supplied, recomputes each
// hash, and compares against the stored ImmutableHash.  The caller is
// responsible for passing entries in insertion order — a typical
// implementation fetches them by timestamp-then-id from the sink.
func VerifyChain(entries []*AuditLog) VerificationResult {
	result := VerificationResult{Entries: len(entries)}
	prev := GenesisHash
	for _, e := range entries {
		expected, err := ComputeHash(e, prev)
		if err != nil {
			return VerificationResult{
				Entries: len(entries),
				Broken:  e,
				Reason:  fmt.Sprintf("hash computation failed: %v", err),
			}
		}
		if e.ImmutableHash != expected {
			return VerificationResult{
				Entries: len(entries),
				Broken:  e,
				Reason: fmt.Sprintf(
					"hash mismatch at id=%s: stored=%s expected=%s",
					e.ID, e.ImmutableHash, expected),
			}
		}
		prev = e.ImmutableHash
	}
	result.OK = true
	return result
}
