package usermanagement

import (
	"sipon-api/internal/app/port"
	userrepo "sipon-api/internal/domain/user/repository"
)

type Dependencies struct {
	UserRepo  userrepo.UserRepository
	ReadModel port.UserQueryReadModel
	Hasher    port.PasswordHasher
}

type UseCases struct {
	ListUsers         *ListUsersUseCase
	GetUser           *GetUserUseCase
	CreateUser        *CreateUserUseCase
	ResetUserPassword *ResetUserPasswordUseCase
	DeactivateUser    *DeactivateUserUseCase
	ReactivateUser    *ReactivateUserUseCase
}

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
