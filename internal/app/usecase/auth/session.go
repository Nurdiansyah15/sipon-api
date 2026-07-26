package authUsecase

import (
	"context"

	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/service/principal"
	userrepo "sipon-api/internal/domain/user/repository"
)

type GetSessionUseCase struct {
	userRepo userrepo.UserRepository
}

func NewGetSessionUseCase(userRepo userrepo.UserRepository) *GetSessionUseCase {
	return &GetSessionUseCase{userRepo: userRepo}
}

// Required — role: any | perm: - | benefit: -
func (uc *GetSessionUseCase) Execute(ctx context.Context, p *principal.Principal) (*dto.SessionData, error) {
	user, err := uc.userRepo.FindByID(ctx, p.UserID)
	if err != nil {
		return nil, err
	}

	name := ""
	if user.Fullname != nil {
		name = *user.Fullname
	}

	roles := make([]dto.SessionRole, 0, len(p.Roles))
	for _, r := range p.Roles {
		roles = append(roles, dto.SessionRole{
			Name:      r.Name,
			RoleType:  r.RoleType,
			ScopeType: r.ScopeType,
			ScopeID:   r.ScopeID,
		})
	}

	permissions := make([]dto.SessionPermission, 0, len(p.Permissions))
	for _, perm := range p.Permissions {
		permissions = append(permissions, dto.SessionPermission{
			Key:   perm.Key,
			Scope: perm.Scope,
		})
	}

	return &dto.SessionData{
		User: dto.SessionUser{
			ID:       user.ID,
			Name:     name,
			Email:    user.Email.Value(),
			Username: user.Username.Value(),
		},
		Roles:       roles,
		Permissions: permissions,
	}, nil
}
