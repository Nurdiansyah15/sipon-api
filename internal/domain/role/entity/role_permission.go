package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/role/constant"
	"strings"
	"time"
)

// RolePermission adalah assignment satu permission ke satu role CUSTOM.
// Hanya berlaku untuk role bertipe custom — permission role system tetap
// dihitung dari constant.RolePermissions (lihat Role.HasPermission).
type RolePermission struct {
	ID            string
	RoleID        string
	PermissionKey constant.PermissionKey
	AssignedAt    time.Time
	AssignedBy    string
	Notes         *string
}

func NewRolePermission(id, roleID string, permissionKey constant.PermissionKey, assignedBy string, notes *string) (*RolePermission, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeRolePermissionKeyRequired)
	}
	if roleID == "" {
		return nil, domainerr.New(constant.CodeRoleIDRequired)
	}
	if strings.TrimSpace(string(permissionKey)) == "" {
		return nil, domainerr.New(constant.CodeRolePermissionKeyRequired)
	}
	if !constant.IsValidPermissionKey(permissionKey) {
		return nil, domainerr.New(constant.CodeRolePermissionKeyInvalid)
	}

	return &RolePermission{
		ID:            id,
		RoleID:        roleID,
		PermissionKey: permissionKey,
		AssignedAt:    time.Now(),
		AssignedBy:    assignedBy,
		Notes:         notes,
	}, nil
}
