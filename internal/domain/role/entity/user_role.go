package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/role/constant"
	"time"
)

// UserRole adalah assignment role ke user pada scope tertentu.
type UserRole struct {
	ID            string
	UserID        string
	RoleID        string
	ScopeType     constant.ScopeType
	ScopeID       *string
	AssignedAt    time.Time
	AssignedBy    string
	ExpiredAt     *time.Time
	IsActive      bool
	Notes         *string // keterangan assignment, diisi dari backoffice
	DeactivatedAt *time.Time
}

func NewUserRole(id, userID, roleID string, scopeType constant.ScopeType, scopeID *string, assignedBy string, expiredAt *time.Time) (*UserRole, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeUserRoleIDRequired)
	}
	if userID == "" {
		return nil, domainerr.New(constant.CodeUserRoleUserIDRequired)
	}
	if roleID == "" {
		return nil, domainerr.New(constant.CodeUserRoleRoleIDRequired)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, domainerr.New(constant.CodeUserRoleScopeTypeInvalid)
	}
	if scopeType == constant.ScopeTypeGlobal && scopeID != nil {
		return nil, domainerr.New(constant.CodeUserRoleScopeIDMustBeEmpty)
	}
	if (scopeType == constant.ScopeTypeRegion || scopeType == constant.ScopeTypeCommunity) && (scopeID == nil || *scopeID == "") {
		return nil, domainerr.New(constant.CodeUserRoleScopeIDRequired)
	}

	return &UserRole{
		ID:         id,
		UserID:     userID,
		RoleID:     roleID,
		ScopeType:  scopeType,
		ScopeID:    scopeID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
		ExpiredAt:  expiredAt,
		IsActive:   true,
	}, nil
}

func (u *UserRole) IsExpired(now time.Time) bool {
	if u.ExpiredAt == nil {
		return false
	}
	return !u.ExpiredAt.After(now)
}

func (u *UserRole) IsUsable(now time.Time) bool {
	return u.IsActive && !u.IsExpired(now)
}

func (u *UserRole) Deactivate(at time.Time) {
	u.IsActive = false
	u.DeactivatedAt = &at
}

func (u *UserRole) Reactivate() {
	u.IsActive = true
	u.DeactivatedAt = nil
}

func (u *UserRole) UpdateExpiration(expiredAt *time.Time) {
	u.ExpiredAt = expiredAt
}
