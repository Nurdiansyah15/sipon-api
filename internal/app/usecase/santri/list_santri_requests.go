package santriUsecase

import (
	"context"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	santrirepo "sipon-api/internal/domain/santri/repository"
	userrepo "sipon-api/internal/domain/user/repository"
)

type ListSantriRequestsUseCase struct {
	santriRequestRepo santrirepo.SantriRequestRepository
	userRepo          userrepo.UserRepository
}

func NewListSantriRequestsUseCase(
	santriRequestRepo santrirepo.SantriRequestRepository,
	userRepo userrepo.UserRepository,
) *ListSantriRequestsUseCase {
	return &ListSantriRequestsUseCase{santriRequestRepo: santriRequestRepo, userRepo: userRepo}
}

func (uc *ListSantriRequestsUseCase) Execute(ctx context.Context, query ListSantriRequestsQuery) ([]SantriRequestItem, dto.Meta, error) {
	q := santrirepo.SantriRequestListQuery{
		Status:   query.Status,
		Page:     query.Page,
		Limit:    query.Limit,
		SortBy:   query.SortBy,
		SortType: query.SortType,
	}

	result, err := uc.santriRequestRepo.List(ctx, q)
	if err != nil {
		return nil, dto.Meta{}, apperror.Internal("gagal mengambil data request", err)
	}

	items := make([]SantriRequestItem, 0, len(result.Items))
	for _, r := range result.Items {
		item := SantriRequestItem{
			ID:        r.ID,
			UserID:    r.UserID,
			Status:    string(r.Status),
			Notes:     r.Notes,
			CreatedAt: r.CreatedAt,
		}

		user, userErr := uc.userRepo.FindByID(ctx, r.UserID)
		if userErr == nil {
			item.Username = user.Username.Value()
			item.Fullname = user.Fullname
			item.Email = user.Email.Value()
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
