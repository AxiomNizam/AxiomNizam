package tenant

// TenantToResponse converts a Tenant to a TenantResponse.
func TenantToResponse(t *Tenant) TenantResponse {
	return TenantResponse{
		ID:             t.ID,
		Name:           t.Name,
		DisplayName:    t.DisplayName,
		Owner:          t.Owner,
		Tier:           t.Tier,
		IsolationLevel: t.IsolationLevel,
		Status:         t.Status,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		Metadata:       t.Metadata,
	}
}
