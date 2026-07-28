package entity

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/valueobject"
)

type UserScope struct {
	ID         string
	UserID     string
	ScopeType  valueobject.UserScopeType
	ScopeValue string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewUserScope(id, userID string, scopeType valueobject.UserScopeType, scopeValue string) (*UserScope, error) {
	if id == "" {
		return nil, domainerr.New("DOMAIN_USER_SCOPE_ID_REQUIRED")
	}
	if userID == "" {
		return nil, domainerr.New("DOMAIN_USER_SCOPE_USER_ID_REQUIRED")
	}

	normalizedValue, err := valueobject.NewUserScopeValue(scopeType, scopeValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &UserScope{
		ID:         id,
		UserID:     userID,
		ScopeType:  scopeType,
		ScopeValue: normalizedValue,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
