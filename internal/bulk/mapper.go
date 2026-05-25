package bulk

// BulkOpToResponse converts a BulkOperation to a BulkOpResponse.
func BulkOpToResponse(op *BulkOperation) BulkOpResponse {
	var progress float64
	if op.TotalItems > 0 {
		progress = float64(op.SuccessCount+op.FailureCount) / float64(op.TotalItems) * 100
	}
	return BulkOpResponse{
		ID:           op.ID,
		TenantID:     op.TenantID,
		Type:         string(op.Type),
		Status:       string(op.Status),
		TotalItems:   op.TotalItems,
		SuccessCount: op.SuccessCount,
		FailureCount: op.FailureCount,
		SkippedCount: op.SkippedCount,
		Progress:     progress,
		CreatedAt:    op.CreatedAt,
		StartedAt:    op.StartedAt,
		CompletedAt:  op.CompletedAt,
	}
}

// BulkOpToProgressResponse converts a BulkOperation to a ProgressResponse.
func BulkOpToProgressResponse(op *BulkOperation) ProgressResponse {
	var progress float64
	if op.TotalItems > 0 {
		progress = float64(op.SuccessCount+op.FailureCount) / float64(op.TotalItems) * 100
	}
	return ProgressResponse{
		ID:       op.ID,
		Status:   string(op.Status),
		Progress: progress,
		Total:    op.TotalItems,
		Success:  op.SuccessCount,
		Failed:   op.FailureCount,
		Skipped:  op.SkippedCount,
	}
}
