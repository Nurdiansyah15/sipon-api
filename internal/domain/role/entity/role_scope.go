package entity

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/role/valueobject"
)

type RoleScope struct {
	ID         string
	RoleID     string
	ScopeType  valueobject.ScopeType
	ScopeValue string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewRoleScope(id, roleID string, scopeType valueobject.ScopeType, scopeValue string) (*RoleScope, error) {
	if id == "" {
		return nil, domainerr.New("DOMAIN_ROLE_SCOPE_ID_REQUIRED")
	}
	if roleID == "" {
		return nil, domainerr.New("DOMAIN_ROLE_SCOPE_ROLE_ID_REQUIRED")
	}

	normalizedValue, err := valueobject.NewScopeValue(scopeType, scopeValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &RoleScope{
		ID:         id,
		RoleID:     roleID,
		ScopeType:  scopeType,
		ScopeValue: normalizedValue,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
