package principal

import (
	"context"
	"errors"
	"strings"

	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type Builder struct {
	userRepo           userrepo.UserRepository
	userRoleRepo       rolerepo.UserRoleRepository
	roleRepo           rolerepo.RoleRepository
	rolePermissionRepo rolerepo.RolePermissionRepository
	userScopeRepo      userrepo.UserScopeRepository
}

func NewBuilder(
	userRepo userrepo.UserRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermissionRepo rolerepo.RolePermissionRepository,
	userScopeRepo userrepo.UserScopeRepository,
) *Builder {
	return &Builder{
		userRepo:           userRepo,
		userRoleRepo:       userRoleRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		userScopeRepo:      userScopeRepo,
	}
}

// Build memuat Principal dari user + role aktifnya. Permission role SYSTEM
// dihitung langsung dari constant.PermissionsForRole(role.Name) (fixed di
// kode); permission role CUSTOM dimuat dari tabel role_permissions (bisa
// diatur dinamis lewat AssignRolePermissionUseCase/RevokeRolePermissionUseCase).
func (b *Builder) Build(ctx context.Context, userID, sessionID string) (*Principal, error) {
	user, err := b.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Status != userconstant.StatusActive {
		return nil, domainerr.New(userconstant.CodeUserNotFound)
	}

	p := &Principal{
		UserID:      userID,
		SessionID:   sessionID,
		Roles:       make([]Role, 0),
		Permissions: make([]Permission, 0),
		Scopes:      make([]UserScope, 0),
	}

	assignments, err := b.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	seenPermKey := make(map[string]struct{})

	for _, assignment := range assignments {
		if assignment == nil {
			continue
		}

		role, roleErr := b.roleRepo.FindByID(ctx, assignment.RoleID)
		if roleErr != nil {
			var de *domainerr.DomainError
			if errors.As(roleErr, &de) && de.Code == roleconstant.CodeRoleNotFound {
				continue
			}
			return nil, roleErr
		}

		p.Roles = append(p.Roles, Role{
			Name:      string(role.Name),
			RoleType:  string(role.RoleType),
			ScopeType: string(assignment.ScopeType),
			ScopeID:   assignment.ScopeID,
		})

		permKeys, permErr := b.effectivePermissionKeys(ctx, role)
		if permErr != nil {
			return nil, permErr
		}
		for _, permKey := range permKeys {
			key := strings.ToLower(string(permKey))
			if _, seen := seenPermKey[key]; seen {
				continue
			}
			seenPermKey[key] = struct{}{}
			p.Permissions = append(p.Permissions, Permission{
				Key:   string(permKey),
				Scope: string(role.ScopeType),
			})
		}
	}

	// Load user_scopes (best-effort — gagal load scope tidak boleh gagalkan request).
	if b.userScopeRepo != nil {
		scopes, scopeErr := b.userScopeRepo.FindByUserID(ctx, userID)
		if scopeErr == nil {
			for _, s := range scopes {
				p.Scopes = append(p.Scopes, UserScope{
					ScopeType:  string(s.ScopeType),
					ScopeValue: s.ScopeValue,
				})
			}
		}
	}

	return p, nil
}

func (b *Builder) effectivePermissionKeys(ctx context.Context, role *roleentity.Role) ([]roleconstant.PermissionKey, error) {
	if role.IsSystem() {
		return roleconstant.PermissionsForRole(role.Name), nil
	}
	assigned, err := b.rolePermissionRepo.ListByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	keys := make([]roleconstant.PermissionKey, 0, len(assigned))
	for _, a := range assigned {
		keys = append(keys, a.PermissionKey)
	}
	return keys, nil
}
