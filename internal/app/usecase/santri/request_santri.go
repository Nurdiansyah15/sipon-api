package santriUsecase

import (
	"context"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/domain/santri/entity"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type RequestSantriUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriRequestRepo santrirepo.SantriRequestRepository
}

func NewRequestSantriUseCase(
	santriRepo santrirepo.SantriRepository,
	santriRequestRepo santrirepo.SantriRequestRepository,
) *RequestSantriUseCase {
	return &RequestSantriUseCase{santriRepo: santriRepo, santriRequestRepo: santriRequestRepo}
}

func (uc *RequestSantriUseCase) Execute(ctx context.Context, userID string) (*RequestSantriResponse, error) {
	_, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err == nil {
		return nil, apperror.Conflict("anda sudah terdaftar sebagai santri")
	}

	existing, _ := uc.santriRequestRepo.FindPendingByUserID(ctx, userID)
	if existing != nil {
		return nil, apperror.Conflict("anda sudah memiliki request yang sedang diproses")
	}

	req, err := entity.NewSantriRequest(uuid.NewString(), userID)
	if err != nil {
		return nil, apperror.Internal("gagal membuat request", err)
	}

	if err := uc.santriRequestRepo.Save(ctx, req); err != nil {
		return nil, apperror.Internal("gagal menyimpan request", err)
	}

	return &RequestSantriResponse{
		ID:      req.ID,
		Message: "request santri berhasil dikirim, menunggu persetujuan admin",
	}, nil
}
