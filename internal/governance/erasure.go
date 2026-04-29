package governance

// =====================================================
// WS-6.1 — Right-to-Erasure Workflow (GDPR Article 17)
//
// Handles data subject erasure requests by scanning all catalog
// assets for the subject's data, executing deletion/anonymization,
// and generating a compliance certificate.
// =====================================================

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ErasureRequest represents a data subject's right-to-erasure request.
type ErasureRequest struct {
	RequestID    string    `json:"requestId"`
	SubjectID    string    `json:"subjectId"`    // Email, user ID, etc.
	SubjectType  string    `json:"subjectType"`  // email, user_id, phone
	RequestedAt  time.Time `json:"requestedAt"`
	RequestedBy  string    `json:"requestedBy"`
	Reason       string    `json:"reason,omitempty"`
}

// ErasureResult records what was done for a single asset.
type ErasureAssetResult struct {
	AssetRef       string `json:"assetRef"`
	Action         string `json:"action"`         // deleted, anonymized, not_found, skipped
	RecordsAffected int64 `json:"recordsAffected"`
	Error          string `json:"error,omitempty"`
}

// ErasureCertificate is the compliance proof of erasure completion.
type ErasureCertificate struct {
	RequestID      string               `json:"requestId"`
	SubjectID      string               `json:"subjectId"`
	CompletedAt    time.Time            `json:"completedAt"`
	AssetsScanned  int                  `json:"assetsScanned"`
	AssetsAffected int                  `json:"assetsAffected"`
	TotalRecords   int64                `json:"totalRecords"`
	Results        []ErasureAssetResult `json:"results"`
	ProofHash      string               `json:"proofHash"` // SHA-256 of results
	Status         string               `json:"status"`    // completed, partial, failed
}

// DataEraser abstracts deleting/anonymizing subject data from a datasource.
type DataEraser interface {
	// FindSubjectData checks if a subject's data exists in an asset.
	FindSubjectData(ctx context.Context, assetRef, subjectID, subjectType string) (int64, error)

	// EraseSubjectData deletes or anonymizes a subject's data from an asset.
	EraseSubjectData(ctx context.Context, assetRef, subjectID, subjectType, action string) (int64, error)
}

// ErasureWorkflow orchestrates the right-to-erasure process.
type ErasureWorkflow struct {
	eraser DataEraser
}

// NewErasureWorkflow creates a new workflow.
func NewErasureWorkflow(eraser DataEraser) *ErasureWorkflow {
	return &ErasureWorkflow{eraser: eraser}
}

// Execute processes an erasure request across all provided assets.
func (w *ErasureWorkflow) Execute(ctx context.Context, req ErasureRequest, assets []AuditableAsset, action string) (*ErasureCertificate, error) {
	if w.eraser == nil {
		return nil, fmt.Errorf("erasure: no data eraser configured")
	}

	if action == "" {
		action = "delete"
	}

	cert := &ErasureCertificate{
		RequestID:     req.RequestID,
		SubjectID:     req.SubjectID,
		AssetsScanned: len(assets),
	}

	var totalRecords int64

	for _, asset := range assets {
		// Respect context cancellation between assets.
		select {
		case <-ctx.Done():
			cert.Status = "cancelled"
			cert.TotalRecords = totalRecords
			cert.CompletedAt = time.Now()
			cert.ProofHash = w.generateProofHash(cert)
			return cert, ctx.Err()
		default:
		}

		// Check if subject data exists.
		count, err := w.eraser.FindSubjectData(ctx, asset.Name, req.SubjectID, req.SubjectType)
		if err != nil {
			cert.Results = append(cert.Results, ErasureAssetResult{
				AssetRef: asset.Name,
				Action:   "error",
				Error:    err.Error(),
			})
			continue
		}

		if count == 0 {
			cert.Results = append(cert.Results, ErasureAssetResult{
				AssetRef: asset.Name,
				Action:   "not_found",
			})
			continue
		}

		// Execute erasure.
		affected, err := w.eraser.EraseSubjectData(ctx, asset.Name, req.SubjectID, req.SubjectType, action)
		if err != nil {
			cert.Results = append(cert.Results, ErasureAssetResult{
				AssetRef:        asset.Name,
				Action:          "error",
				RecordsAffected: 0,
				Error:           err.Error(),
			})
			continue
		}

		cert.Results = append(cert.Results, ErasureAssetResult{
			AssetRef:        asset.Name,
			Action:          action,
			RecordsAffected: affected,
		})
		cert.AssetsAffected++
		totalRecords += affected
	}

	cert.TotalRecords = totalRecords
	cert.CompletedAt = time.Now()

	// Determine status.
	hasErrors := false
	for _, r := range cert.Results {
		if r.Action == "error" {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		cert.Status = "partial"
	} else {
		cert.Status = "completed"
	}

	// Generate proof hash.
	cert.ProofHash = w.generateProofHash(cert)

	return cert, nil
}

// generateProofHash creates a SHA-256 hash of the erasure results for tamper detection.
func (w *ErasureWorkflow) generateProofHash(cert *ErasureCertificate) string {
	data := fmt.Sprintf("%s|%s|%s|%d|%d",
		cert.RequestID, cert.SubjectID, cert.CompletedAt.Format(time.RFC3339),
		cert.AssetsScanned, cert.TotalRecords)

	for _, r := range cert.Results {
		data += fmt.Sprintf("|%s:%s:%d", r.AssetRef, r.Action, r.RecordsAffected)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
