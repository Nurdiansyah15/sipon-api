package repository

import (
	"context"

	"sipon-api/internal/domain/santri/entity"
)

type SantriListResult struct {
	Items []*entity.Santri
	Meta  SantriListMeta
}

type SantriListMeta struct {
	CurrentPage int64
	PerPage     int64
	Total       int64
	TotalPages  int64
}

type SantriRequestListResult struct {
	Items []*entity.SantriRequest
	Meta  SantriListMeta
}

type SantriRepository interface {
	Save(ctx context.Context, santri *entity.Santri) error
	Update(ctx context.Context, santri *entity.Santri) error
	FindByID(ctx context.Context, id string) (*entity.Santri, error)
	FindByUserID(ctx context.Context, userID string) (*entity.Santri, error)
	FindByNIS(ctx context.Context, nis string) (*entity.Santri, error)
	List(ctx context.Context, query SantriListQuery) (*SantriListResult, error)
}

type SantriListQuery struct {
	NIS      *string
	UserID   *string
	Page     *int
	Limit    *int
	SortBy   *string
	SortType *string
}

type SantriDokumenRepository interface {
	Save(ctx context.Context, dokumen *entity.SantriDokumen) error
	Update(ctx context.Context, dokumen *entity.SantriDokumen) error
	FindByID(ctx context.Context, id string) (*entity.SantriDokumen, error)
	FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriDokumen, error)
	FindBySantriIDAndKind(ctx context.Context, santriID, kind string) ([]*entity.SantriDokumen, error)
	Delete(ctx context.Context, id string) error
}

type SantriRequestRepository interface {
	Save(ctx context.Context, req *entity.SantriRequest) error
	Update(ctx context.Context, req *entity.SantriRequest) error
	FindByID(ctx context.Context, id string) (*entity.SantriRequest, error)
	FindPendingByUserID(ctx context.Context, userID string) (*entity.SantriRequest, error)
	FindByStatus(ctx context.Context, status string) ([]*entity.SantriRequest, error)
	List(ctx context.Context, query SantriRequestListQuery) (*SantriRequestListResult, error)
}

type SantriRequestListQuery struct {
	Status   *string
	Page     *int
	Limit    *int
	SortBy   *string
	SortType *string
}
