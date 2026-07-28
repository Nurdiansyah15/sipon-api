package rolepermission

import (
	"sipon-api/internal/app/port"
	rolerepo "sipon-api/internal/domain/role/repository"
	userrepo "sipon-api/internal/domain/user/repository"
)

type Dependencies struct {
	RoleRepo           rolerepo.RoleRepository
	UserRoleRepo       rolerepo.UserRoleRepository
	RolePermissionRepo rolerepo.RolePermissionRepository
	RoleScopeRepo      rolerepo.RoleScopeRepository
	UserRepo           userrepo.UserRepository
	ReadModel          port.RolePermissionQueryReadModel
}

type UseCases struct {
	ListRoles            *ListRolesUseCase
	GetRole              *GetRoleUseCase
	CreateRole           *CreateRoleUseCase
	UpdateRole           *UpdateRoleUseCase
	AssignRolePermission *AssignRolePermissionUseCase
	RevokeRolePermission *RevokeRolePermissionUseCase
	ListPermissionKeys   *ListPermissionKeysUseCase
	ListUserRoles        *ListUserRolesUseCase
	GetUserRole          *GetUserRoleUseCase
	AssignUserRole       *AssignUserRoleUseCase
	UpdateUserRole       *UpdateUserRoleUseCase
	DeactivateUserRole   *DeactivateUserRoleUseCase
	ReactivateUserRole   *ReactivateUserRoleUseCase
	DeleteUserRole       *DeleteUserRoleUseCase
	AssignRoleScope      *AssignRoleScopeUseCase
	RemoveRoleScope      *RemoveRoleScopeUseCase
	ListRoleScopes       *ListRoleScopesUseCase
}

func NewUseCases(deps Dependencies) *UseCases {
	newGetUserRole := func() *GetUserRoleUseCase {
		return NewGetUserRoleUseCase(deps.UserRepo, deps.RoleRepo, deps.UserRoleRepo, deps.RolePermissionRepo)
	}
	return &UseCases{
		ListRoles:            NewListRolesUseCase(deps.ReadModel),
		GetRole:              NewGetRoleUseCase(deps.RoleRepo, deps.RolePermissionRepo),
		CreateRole:           NewCreateRoleUseCase(deps.RoleRepo, deps.RolePermissionRepo),
		UpdateRole:           NewUpdateRoleUseCase(deps.RoleRepo, deps.RolePermissionRepo),
		AssignRolePermission: NewAssignRolePermissionUseCase(deps.RoleRepo, deps.RolePermissionRepo),
		RevokeRolePermission: NewRevokeRolePermissionUseCase(deps.RoleRepo, deps.RolePermissionRepo),
		ListPermissionKeys:   NewListPermissionKeysUseCase(),
		ListUserRoles:        NewListUserRolesUseCase(deps.ReadModel),
		GetUserRole:          newGetUserRole(),
		AssignUserRole:       NewAssignUserRoleUseCase(deps.RoleRepo, deps.UserRoleRepo, newGetUserRole()),
		UpdateUserRole:       NewUpdateUserRoleUseCase(deps.UserRoleRepo, newGetUserRole()),
		DeactivateUserRole:   NewDeactivateUserRoleUseCase(deps.UserRoleRepo, newGetUserRole()),
		ReactivateUserRole:   NewReactivateUserRoleUseCase(deps.UserRoleRepo, newGetUserRole()),
		DeleteUserRole:       NewDeleteUserRoleUseCase(deps.UserRoleRepo),
		AssignRoleScope:      NewAssignRoleScopeUseCase(deps.RoleRepo, deps.RoleScopeRepo),
		RemoveRoleScope:      NewRemoveRoleScopeUseCase(deps.RoleScopeRepo),
		ListRoleScopes:       NewListRoleScopesUseCase(deps.RoleScopeRepo),
	}
}
