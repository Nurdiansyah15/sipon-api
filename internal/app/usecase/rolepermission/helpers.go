package rolepermission

import (
	"context"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"
	userrepo "sipon-api/internal/domain/user/repository"

	"github.com/google/uuid"
)

func parseRoleType(raw string, required bool) (roleconstant.RoleType, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return "", apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}
		return "", nil
	}
	roleType := roleconstant.RoleType(raw)
	switch roleType {
	case roleconstant.RoleTypeSystem, roleconstant.RoleTypeCustom:
		return roleType, nil
	default:
		return "", apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}
}

func parseScopeType(raw string, required bool) (roleconstant.ScopeType, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return "", apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
		}
		return "", nil
	}
	scopeType := roleconstant.ScopeType(raw)
	switch scopeType {
	case roleconstant.ScopeTypeGlobal, roleconstant.ScopeTypeRegion, roleconstant.ScopeTypeCommunity:
		return scopeType, nil
	default:
		return "", apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}
}

func parseOptionalBool(raw string) (*bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	switch strings.ToLower(raw) {
	case "true":
		value := true
		return &value, nil
	case "false":
		value := false
		return &value, nil
	default:
		return nil, apperror.Unprocessable(string(apperror.CodeUnprocessable), nil)
	}
}

func permissionKeyStrings(keys []roleconstant.PermissionKey) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, string(k))
	}
	return out
}

func mapRoleSummaryReadItem(item port.RoleSummaryReadItem) dto.RoleSummaryResponse {
	return dto.RoleSummaryResponse{ID: item.ID, Name: item.Name, DisplayName: item.DisplayName, RoleType: item.RoleType, Assignable: item.Assignable}
}

func mapRoleEntitySummary(item *roleentity.Role) dto.RoleSummaryResponse {
	return dto.RoleSummaryResponse{ID: item.ID, Name: string(item.Name), DisplayName: item.DisplayName, RoleType: string(item.RoleType), Assignable: item.Assignable}
}

// mapRoleReadItem dipakai oleh ListRoles. Permission role CUSTOM sengaja TIDAK
// disertakan di sini (butuh query role_permissions per baris/N+1) — ambil
// detail lewat GET /role-permission/roles/:role_id (buildRoleResponse) untuk
// permission lengkap. Role system tetap menampilkan permission-nya karena
// gratis diambil dari constant, tanpa query tambahan.
func mapRoleReadItem(item port.RoleReadItem) dto.RoleResponse {
	var permissions []string
	if item.RoleType == string(roleconstant.RoleTypeSystem) {
		permissions = permissionKeyStrings(roleconstant.PermissionsForRole(roleconstant.RoleName(item.Name)))
	}
	return dto.RoleResponse{
		ID: item.ID, Name: item.Name, DisplayName: item.DisplayName, Description: item.Description,
		RoleType: item.RoleType, ScopeType: item.ScopeType, Assignable: item.Assignable,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		Permissions: permissions,
	}
}

func mapUserRoleReadItem(item port.UserRoleReadItem) dto.UserRoleResponse {
	return dto.UserRoleResponse{
		ID: item.ID, UserID: item.UserID,
		User:          dto.UserSummaryResponse{ID: item.User.ID, Name: item.User.Name, Email: item.User.Email, Phone: item.User.Phone},
		RoleID:        item.RoleID,
		Role:          mapRoleSummaryReadItem(item.Role),
		ScopeType:     item.ScopeType,
		ScopeID:       item.ScopeID,
		AssignedAt:    item.AssignedAt,
		AssignedBy:    item.AssignedBy,
		ExpiredAt:     item.ExpiredAt,
		IsActive:      item.IsActive,
		DeactivatedAt: item.DeactivatedAt,
		Permissions:   item.Permissions,
	}
}

func mapRoleDomainValidationError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		return apperror.Unprocessable(string(de.Code), nil, err)
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

func mapRoleDomainError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case roleconstant.CodeRoleNotFound:
			return apperror.NotFound(string(de.Code), err)
		case roleconstant.CodeRoleDuplicateName:
			return apperror.Conflict(string(de.Code), err)
		case roleconstant.CodeRoleQueryFailed, roleconstant.CodeRolePersistenceFailed:
			return apperror.Internal(string(apperror.CodeInternal), err)
		case roleconstant.CodeRoleNotAssignable:
			return apperror.Conflict(string(de.Code), err)
		default:
			return apperror.Unprocessable(string(de.Code), nil, err)
		}
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

func mapUserRoleDomainError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case roleconstant.CodeUserRoleNotFound:
			return apperror.NotFound(string(de.Code), err)
		case roleconstant.CodeUserRoleDuplicate:
			return apperror.Conflict(string(de.Code), err)
		case roleconstant.CodeUserRoleQueryFailed, roleconstant.CodeUserRolePersistenceFailed:
			return apperror.Internal(string(apperror.CodeInternal), err)
		default:
			return apperror.Unprocessable(string(de.Code), nil, err)
		}
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

func mapRolePermissionDomainError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case roleconstant.CodeRolePermissionNotFound:
			return apperror.NotFound(string(de.Code), err)
		case roleconstant.CodeRolePermissionDuplicate:
			return apperror.Conflict(string(de.Code), err)
		case roleconstant.CodeRolePermissionRequiresCustom:
			return apperror.Conflict(string(de.Code), err)
		case roleconstant.CodeRolePermissionQueryFailed, roleconstant.CodeRolePermissionPersistenceFailed:
			return apperror.Internal(string(apperror.CodeInternal), err)
		default:
			return apperror.Unprocessable(string(de.Code), nil, err)
		}
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

// resolveRolePermissions mengembalikan permission key efektif suatu role:
// role system → constant.RolePermissions (fixed, tidak bisa diubah lewat API);
// role custom → tabel role_permissions (bisa diatur dinamis via
// AssignRolePermissionUseCase/RevokeRolePermissionUseCase).
func resolveRolePermissions(ctx context.Context, rolePermissionRepo rolerepo.RolePermissionRepository, role *roleentity.Role) ([]string, error) {
	if role.IsSystem() {
		return permissionKeyStrings(roleconstant.PermissionsForRole(role.Name)), nil
	}
	assigned, err := rolePermissionRepo.ListByRoleID(ctx, role.ID)
	if err != nil {
		return nil, mapRolePermissionDomainError(err)
	}
	keys := make([]string, 0, len(assigned))
	for _, a := range assigned {
		keys = append(keys, string(a.PermissionKey))
	}
	return keys, nil
}

func buildRoleResponse(ctx context.Context, roleRepo rolerepo.RoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository, roleID string) (*dto.RoleResponse, error) {
	role, err := roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, mapRoleDomainError(err)
	}
	permissions, err := resolveRolePermissions(ctx, rolePermissionRepo, role)
	if err != nil {
		return nil, err
	}
	resp := &dto.RoleResponse{
		ID: role.ID, Name: string(role.Name), DisplayName: role.DisplayName, Description: role.Description,
		RoleType: string(role.RoleType), ScopeType: string(role.ScopeType), Assignable: role.Assignable,
		CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
		Permissions: permissions,
	}
	return resp, nil
}

func buildUserRoleResponse(ctx context.Context, userRepo userrepo.UserRepository, roleRepo rolerepo.RoleRepository, userRoleRepo rolerepo.UserRoleRepository, rolePermissionRepo rolerepo.RolePermissionRepository, userRoleID string) (*dto.UserRoleResponse, error) {
	assignment, err := userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, mapUserRoleDomainError(err)
	}
	user, err := userRepo.FindByID(ctx, assignment.UserID)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	role, err := roleRepo.FindByID(ctx, assignment.RoleID)
	if err != nil {
		return nil, mapRoleDomainError(err)
	}
	permissions, err := resolveRolePermissions(ctx, rolePermissionRepo, role)
	if err != nil {
		return nil, err
	}
	var email *string
	if value := strings.TrimSpace(user.Email.Value()); value != "" {
		email = &value
	}
	var phone *string
	if user.PhoneNumber != nil {
		value := user.PhoneNumber.Value()
		phone = &value
	}
	resp := &dto.UserRoleResponse{
		ID: assignment.ID, UserID: assignment.UserID,
		User:          dto.UserSummaryResponse{ID: user.ID, Name: user.Fullname, Email: email, Phone: phone},
		RoleID:        assignment.RoleID,
		Role:          mapRoleEntitySummary(role),
		ScopeType:     string(assignment.ScopeType),
		ScopeID:       assignment.ScopeID,
		AssignedAt:    assignment.AssignedAt,
		ExpiredAt:     assignment.ExpiredAt,
		IsActive:      assignment.IsActive,
		DeactivatedAt: assignment.DeactivatedAt,
		Permissions:   permissions,
	}
	if strings.TrimSpace(assignment.AssignedBy) != "" {
		resp.AssignedBy = &assignment.AssignedBy
	}
	return resp, nil
}

func newRoleEntity(req dto.CreateRoleRequest) (*roleentity.Role, error) {
	roleType, err := parseRoleType(req.RoleType, true)
	if err != nil {
		return nil, err
	}
	scopeType, err := parseScopeType(req.ScopeType, true)
	if err != nil {
		return nil, err
	}
	role, err := roleentity.NewRole(uuid.NewString(), roleconstant.RoleName(strings.TrimSpace(req.Name)), strings.TrimSpace(req.DisplayName), roleType, scopeType, req.Assignable, req.Description)
	if err != nil {
		return nil, mapRoleDomainValidationError(err)
	}
	return role, nil
}

func newUserRoleAssignment(actorUserID string, req dto.AssignUserRoleRequest, roleID string) (*roleentity.UserRole, error) {
	scopeType, err := parseScopeType(req.ScopeType, true)
	if err != nil {
		return nil, err
	}
	assignment, err := roleentity.NewUserRole(uuid.NewString(), strings.TrimSpace(req.UserID), roleID, scopeType, req.ScopeID, strings.TrimSpace(actorUserID), req.ExpiredAt)
	if err != nil {
		return nil, mapRoleDomainValidationError(err)
	}
	return assignment, nil
}
