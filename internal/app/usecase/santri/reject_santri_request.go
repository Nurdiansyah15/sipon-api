package santriUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
)

type RejectSantriRequestUseCase struct {
	santriRequestRepo santrirepo.SantriRequestRepository
}

func NewRejectSantriRequestUseCase(
	santriRequestRepo santrirepo.SantriRequestRepository,
) *RejectSantriRequestUseCase {
	return &RejectSantriRequestUseCase{santriRequestRepo: santriRequestRepo}
}

func (uc *RejectSantriRequestUseCase) Execute(ctx context.Context, requestID, reviewerID string, req RejectSantriRequestRequest) error {
	sr, err := uc.santriRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriRequestNotFound {
			return apperror.NotFound("request tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil request", err)
	}

	if err := sr.Reject(reviewerID, req.Notes); err != nil {
		return apperror.Conflict("request tidak dalam status pending", err)
	}

	if err := uc.santriRequestRepo.Update(ctx, sr); err != nil {
		return apperror.Internal("gagal update request", err)
	}

	return nil
}
