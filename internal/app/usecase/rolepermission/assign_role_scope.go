package rolepermission

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"
	"sipon-api/internal/domain/role/valueobject"
)

type AssignRoleScopeUseCase struct {
	roleRepo      rolerepo.RoleRepository
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewAssignRoleScopeUseCase(
	roleRepo rolerepo.RoleRepository,
	roleScopeRepo rolerepo.RoleScopeRepository,
) *AssignRoleScopeUseCase {
	return &AssignRoleScopeUseCase{roleRepo: roleRepo, roleScopeRepo: roleScopeRepo}
}

func (uc *AssignRoleScopeUseCase) Execute(ctx context.Context, roleID string, req dto.AssignRoleScopeRequest) (*dto.RoleScopeResponse, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, apperror.Unprocessable("role_id wajib diisi", nil)
	}

	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, apperror.NotFound("role tidak ditemukan", err)
	}
	if role.IsSystem() {
		return nil, apperror.Forbidden("system role tidak dapat diberikan scope")
	}

	scopeType := valueobject.ScopeType(strings.TrimSpace(strings.ToLower(req.ScopeType)))
	normalizedValue, err := valueobject.NewScopeValue(scopeType, strings.TrimSpace(strings.ToLower(req.ScopeValue)))
	if err != nil {
		return nil, apperror.Unprocessable("scope value tidak valid", nil, err)
	}

	scope, err := entity.NewRoleScope(uuid.NewString(), roleID, scopeType, normalizedValue)
	if err != nil {
		return nil, apperror.Unprocessable("gagal membuat role scope", nil, err)
	}

	// Cek duplikasi — kombinasi scope_type + scope_value yang sama tidak boleh ada
	existingScopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err == nil {
		for _, es := range existingScopes {
			if es.ScopeType == scopeType && es.ScopeValue == normalizedValue {
				return nil, apperror.Conflict("scope sudah ada pada role ini")
			}
		}
	}

	if err := uc.roleScopeRepo.Save(ctx, scope); err != nil {
		return nil, apperror.Internal("gagal menyimpan role scope", err)
	}

	return &dto.RoleScopeResponse{
		ID: scope.ID, ScopeType: string(scope.ScopeType), ScopeValue: scope.ScopeValue,
	}, nil
}
