package santriUsecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/entity"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
	santriVO "sipon-api/internal/domain/santri/valueobject"
	userentity "sipon-api/internal/domain/user/entity"
	userconst "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
)

type ApproveSantriRequestUseCase struct {
	santriRepo        santrirepo.SantriRepository
	santriRequestRepo santrirepo.SantriRequestRepository
	userRepo          userrepo.UserRepository
}

func NewApproveSantriRequestUseCase(
	santriRepo santrirepo.SantriRepository,
	santriRequestRepo santrirepo.SantriRequestRepository,
	userRepo userrepo.UserRepository,
) *ApproveSantriRequestUseCase {
	return &ApproveSantriRequestUseCase{santriRepo: santriRepo, santriRequestRepo: santriRequestRepo, userRepo: userRepo}
}

func (uc *ApproveSantriRequestUseCase) Execute(ctx context.Context, requestID, reviewerID string, req ApproveSantriRequestRequest) error {
	nis, err := santriVO.NewNIS(req.NIS)
	if err != nil {
		return apperror.Unprocessable("format NIS tidak valid", nil)
	}

	existing, _ := uc.santriRepo.FindByNIS(ctx, nis.Value())
	if existing != nil {
		return apperror.Conflict("NIS sudah digunakan")
	}

	sr, err := uc.santriRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == santriconstant.CodeSantriRequestNotFound {
			return apperror.NotFound("request tidak ditemukan", err)
		}
		return apperror.Internal("gagal mengambil request", err)
	}

	if err := sr.Approve(reviewerID, nis.Value()); err != nil {
		return apperror.Conflict("request tidak dalam status pending", err)
	}

	if err := uc.santriRequestRepo.Update(ctx, sr); err != nil {
		return apperror.Internal("gagal update request", err)
	}

	santri, err := entity.NewSantri(uuid.NewString(), sr.UserID)
	if err != nil {
		return apperror.Internal("gagal membuat santri", err)
	}
	santri.NIS = &nis
	gender := nis.Gender()
	santri.Option = gender

	if err := uc.santriRepo.Save(ctx, santri); err != nil {
		return apperror.Internal("gagal menyimpan santri", err)
	}

	user, err := uc.userRepo.FindByID(ctx, sr.UserID)
	if err == nil {
		cred := user.FindCredential(userconst.CredentialTypeLocal)
		if cred != nil {
			nisIdentity, idErr := userentity.NewLoginIdentity(
				uuid.NewString(), user.ID, cred.ID,
				userconst.LoginIdentifierNIS, nis.Value(), false, nil,
			)
			if idErr == nil {
				cred.AddLoginIdentity(nisIdentity)
				_ = uc.userRepo.Update(ctx, user)
			}
		}
	}

	return nil
}
