package usermanagement

import (
	"sipon-api/internal/app/port"
	userrepo "sipon-api/internal/domain/user/repository"
)

// Dependencies untuk paket usermanagement. Berbeda dari rolepermission
// (memberakses User aggregate, bukan role_permission), tapi memakai UserQueryReadModel
// (read model) untuk listing + role summary — bayangan rolepermission.Dependencies.
type Dependencies struct {
	UserRepo  userrepo.UserRepository
	ReadModel port.UserQueryReadModel
	Hasher    port.PasswordHasher
}

// UseCases mengumpulkan seluruh usecase admin user-management.
type UseCases struct {
	ListUsers          *ListUsersUseCase
	GetUser            *GetUserUseCase
	CreateUser         *CreateUserUseCase
	ResetUserPassword  *ResetUserPasswordUseCase
	DeactivateUser     *DeactivateUserUseCase
	ReactivateUser     *ReactivateUserUseCase
}

// NewUseCases membangun seluruh usecase user-management dari Dependencies.
func NewUseCases(deps Dependencies) *UseCases {
	return &UseCases{
		ListUsers:         NewListUsersUseCase(deps.ReadModel),
		GetUser:          NewGetUserUseCase(deps.UserRepo, deps.ReadModel),
		CreateUser:       NewCreateUserUseCase(deps.UserRepo, deps.Hasher),
		ResetUserPassword: NewResetUserPasswordUseCase(deps.UserRepo, deps.Hasher),
		DeactivateUser:  NewDeactivateUserUseCase(deps.UserRepo),
		ReactivateUser:  NewReactivateUserUseCase(deps.UserRepo),
	}
}