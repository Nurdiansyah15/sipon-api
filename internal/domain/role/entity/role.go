package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/role/constant"
	"strings"
	"time"
)

type Role struct {
	ID          string
	Name        constant.RoleName
	DisplayName string
	Description *string
	RoleType    constant.RoleType
	ScopeType   constant.ScopeType
	Assignable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRole(id string, name constant.RoleName, displayName string, roleType constant.RoleType, scopeType constant.ScopeType, assignable bool, description *string) (*Role, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeRoleIDRequired)
	}
	if name == "" {
		return nil, domainerr.New(constant.CodeRoleNameRequired)
	}
	if displayName == "" {
		return nil, domainerr.New(constant.CodeRoleDisplayNameRequired)
	}
	if roleType != constant.RoleTypeSystem && roleType != constant.RoleTypeCustom {
		return nil, domainerr.New(constant.CodeRoleTypeInvalid)
	}
	if scopeType != constant.ScopeTypeGlobal && scopeType != constant.ScopeTypeRegion && scopeType != constant.ScopeTypeCommunity {
		return nil, domainerr.New(constant.CodeRoleScopeTypeInvalid)
	}

	now := time.Now()
	return &Role{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		RoleType:    roleType,
		ScopeType:   scopeType,
		Assignable:  assignable,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (r *Role) IsSystem() bool {
	return r.RoleType == constant.RoleTypeSystem
}

func (r *Role) EnsureCanDelete() error {
	if r.IsSystem() {
		return domainerr.New(constant.CodeRoleIsSystem)
	}
	return nil
}

func (r *Role) EnsureAssignable() error {
	if r.Assignable {
		return nil
	}
	return domainerr.New(constant.CodeRoleNotAssignable)
}

// EnsureCustom memastikan role ini bertipe custom — dipakai sebelum
// assign/revoke permission via role_permissions, karena permission role
// system fixed dari constant.RolePermissions dan tidak boleh diubah lewat API.
func (r *Role) EnsureCustom() error {
	if r.IsSystem() {
		return domainerr.New(constant.CodeRolePermissionRequiresCustom)
	}
	return nil
}

// EnsureAssignmentScopeMatch memastikan scope_type yang akan di-assign ke user sesuai
// dengan scope_type role. global hanya boleh di-assign ke scope global, region ke region,
// community ke community.
func (r *Role) EnsureAssignmentScopeMatch(assignedScopeType constant.ScopeType) error {
	if r.ScopeType != assignedScopeType {
		return domainerr.New(constant.CodeRoleAssignmentScopeMismatch)
	}
	return nil
}

// HasPermission mengecek permission role SYSTEM ini berdasarkan mapping
// constant.RolePermissions. Hanya berlaku untuk role system — role custom
// selalu mengembalikan false di sini karena permission-nya disimpan di tabel
// role_permissions, bukan di constant (lihat RolePermissionRepository dan
// rolepermission.buildRoleResponse untuk cara resolve permission role custom).
func (r *Role) HasPermission(key constant.PermissionKey) bool {
	return constant.RoleHasPermission(r.Name, key)
}

func (r *Role) Touch() {
	r.UpdatedAt = time.Now()
}

func (r *Role) UpdateDetails(displayName *string, description *string, assignable *bool) error {
	if displayName != nil {
		trimmed := strings.TrimSpace(*displayName)
		if trimmed == "" {
			return domainerr.New(constant.CodeRoleDisplayNameRequired)
		}
		r.DisplayName = trimmed
	}
	if description != nil {
		r.Description = description
	}
	if assignable != nil {
		r.Assignable = *assignable
	}
	r.Touch()
	return nil
}
