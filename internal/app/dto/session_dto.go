package dto

// SessionUser is the user slice of the session bootstrap response.
type SessionUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// SessionRole is one entry in roles[].
type SessionRole struct {
	Name      string  `json:"name"`
	RoleType  string  `json:"role_type"`
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id"`
}

// SessionPermission is one entry in permissions[].
type SessionPermission struct {
	Key   string `json:"key"`
	Scope string `json:"scope"`
}

// SessionUserScope is one entry in scopes[].
type SessionUserScope struct {
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}

// SessionData is the full response body for GET /api/v1/auth/session.
type SessionData struct {
	User        SessionUser         `json:"user"`
	Roles       []SessionRole       `json:"roles"`
	Permissions []SessionPermission `json:"permissions"`
	Scopes      []SessionUserScope  `json:"scopes"`
}
