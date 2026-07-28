package dto

// ── Assign User Scope ────────────────────────────────────────────────────────

type AssignUserScopeRequest struct {
	ScopeType  string `json:"scope_type" binding:"required,oneof=gender"`
	ScopeValue string `json:"scope_value" binding:"required,oneof=male female"`
}

type UserScopeResponse struct {
	ID         string `json:"id"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}
