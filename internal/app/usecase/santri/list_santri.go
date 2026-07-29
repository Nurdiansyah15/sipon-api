package santriUsecase

import (
	"context"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	santrirepo "sipon-api/internal/domain/santri/repository"
	userrepo "sipon-api/internal/domain/user/repository"
)

type ListSantriUseCase struct {
	santriRepo santrirepo.SantriRepository
	userRepo   userrepo.UserRepository
}

func NewListSantriUseCase(
	santriRepo santrirepo.SantriRepository,
	userRepo userrepo.UserRepository,
) *ListSantriUseCase {
	return &ListSantriUseCase{santriRepo: santriRepo, userRepo: userRepo}
}

func (uc *ListSantriUseCase) Execute(ctx context.Context, query ListSantriQuery) ([]ListSantriItem, dto.Meta, error) {
	q := santrirepo.SantriListQuery{
		NIS:      query.NIS,
		Page:     query.Page,
		Limit:    query.Limit,
		SortBy:   query.SortBy,
		SortType: query.SortType,
	}

	result, err := uc.santriRepo.List(ctx, q)
	if err != nil {
		return nil, dto.Meta{}, apperror.Internal("gagal mengambil data santri", err)
	}

	items := make([]ListSantriItem, 0, len(result.Items))
	for _, s := range result.Items {
		item := ListSantriItem{
			ID:        s.ID,
			UserID:    s.UserID,
			CreatedAt: s.CreatedAt,
		}
		if s.NIS != nil {
			v := s.NIS.Value()
			item.NIS = &v
		}

		user, userErr := uc.userRepo.FindByID(ctx, s.UserID)
		if userErr == nil {
			item.Fullname = user.Fullname
			item.Username = user.Username.Value()
			item.Email = user.Email.Value()
			item.Status = string(user.Status)
		}

		items = append(items, item)
	}

	meta := dto.Meta{
		CurrentPage: result.Meta.CurrentPage,
		PerPage:     result.Meta.PerPage,
		Total:       result.Meta.Total,
		TotalPages:  result.Meta.TotalPages,
	}

	return items, meta, nil
}
