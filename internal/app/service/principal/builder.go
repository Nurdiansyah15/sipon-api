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
	roleScopeRepo      rolerepo.RoleScopeRepository
}

func NewBuilder(
	userRepo userrepo.UserRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermissionRepo rolerepo.RolePermissionRepository,
	roleScopeRepo rolerepo.RoleScopeRepository,
) *Builder {
	return &Builder{
		userRepo:           userRepo,
		userRoleRepo:       userRoleRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		roleScopeRepo:      roleScopeRepo,
	}
}

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
	seenScopeKey := make(map[string]struct{})

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

		// Load role_scopes (best-effort, per role)
		if b.roleScopeRepo != nil {
			roleScopes, scopeErr := b.roleScopeRepo.FindByRoleID(ctx, assignment.RoleID)
			if scopeErr == nil {
				for _, rs := range roleScopes {
					dedupKey := string(rs.ScopeType) + ":" + rs.ScopeValue
					if _, seen := seenScopeKey[dedupKey]; seen {
						continue
					}
					seenScopeKey[dedupKey] = struct{}{}
					p.Scopes = append(p.Scopes, UserScope{
						ScopeType:  string(rs.ScopeType),
						ScopeValue: rs.ScopeValue,
					})
				}
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
