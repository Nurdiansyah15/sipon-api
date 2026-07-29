package santriUsecase

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/entity"
	santriconstant "sipon-api/internal/domain/santri/constant"
	santrirepo "sipon-api/internal/domain/santri/repository"
	santriVO "sipon-api/internal/domain/santri/valueobject"
	userconst "sipon-api/internal/domain/user/constant"
	userentity "sipon-api/internal/domain/user/entity"
	userrepo "sipon-api/internal/domain/user/repository"
	uservo "sipon-api/internal/domain/user/valueobject"
)

type CreateSantriUseCase struct {
	userRepo    userrepo.UserRepository
	santriRepo  santrirepo.SantriRepository
	hasher      port.PasswordHasher
}

func NewCreateSantriUseCase(
	userRepo userrepo.UserRepository,
	santriRepo santrirepo.SantriRepository,
	hasher port.PasswordHasher,
) *CreateSantriUseCase {
	return &CreateSantriUseCase{userRepo: userRepo, santriRepo: santriRepo, hasher: hasher}
}

func (uc *CreateSantriUseCase) Execute(ctx context.Context, req CreateSantriRequest) (*CreateSantriResponse, error) {
	nis, err := santriVO.NewNIS(req.NIS)
	if err != nil {
		return nil, apperror.Unprocessable("format NIS tidak valid: harus 1000[12] diikuti 5 digit", nil)
	}

	existing, _ := uc.santriRepo.FindByNIS(ctx, nis.Value())
	if existing != nil {
		return nil, apperror.Conflict("NIS sudah digunakan")
	}

	username, _ := uservo.NewUsername(nis.Value())
	email, _ := uservo.NewEmail(nis.Value() + "@santri.sipon")

	password, err := generateRandomPassword(12)
	if err != nil {
		return nil, apperror.Internal("gagal generate password", err)
	}

	hashedPassword, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, apperror.Internal("gagal hash password", err)
	}
	hashedVO, _ := uservo.NewHashedPassword(hashedPassword)

	nisVal := nis.Value()
	fullname := &nisVal

	user := &userentity.User{
		ID:       uuid.NewString(),
		Username: username,
		Fullname: fullname,
		Email:    email,
		Status:   userconst.StatusActive,
	}

	cred := userentity.NewLocalCredentialWithoutPassword(uuid.NewString(), user.ID, true)
	cred.SecretHash = &hashedVO
	user.AddCredential(cred)

	nisIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), user.ID, cred.ID, userconst.LoginIdentifierNIS, nis.Value(), true, nil)
	if err != nil {
		return nil, apperror.Internal("gagal membuat identitas NIS", err)
	}
	cred.AddLoginIdentity(nisIdentity)

	emailIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), user.ID, cred.ID, userconst.LoginIdentifierEmail, email.Value(), false, nil)
	if err != nil {
		return nil, apperror.Internal("gagal membuat identitas email", err)
	}
	cred.AddLoginIdentity(emailIdentity)

	usernameIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), user.ID, cred.ID, userconst.LoginIdentifierUsername, username.Value(), false, nil)
	if err != nil {
		return nil, apperror.Internal("gagal membuat identitas username", err)
	}
	cred.AddLoginIdentity(usernameIdentity)

	if err := uc.userRepo.Save(ctx, user); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) && de.Code == userconst.CodeUserDuplicate {
			return nil, apperror.Conflict("user dengan identitas tersebut sudah ada", err)
		}
		return nil, apperror.Internal("gagal membuat user", err)
	}

	gender := nis.Gender()
	santri, err := entity.NewSantri(uuid.NewString(), user.ID)
	if err != nil {
		return nil, apperror.Internal("gagal membuat santri", err)
	}
	santri.NIS = &nis
	santri.Option = gender

	if err := uc.santriRepo.Save(ctx, santri); err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case santriconstant.CodeSantriDuplicate:
				return nil, apperror.Conflict("santri duplicate", err)
			}
		}
		return nil, apperror.Internal("gagal menyimpan santri", err)
	}

	return &CreateSantriResponse{
		UserID:            user.ID,
		SantriID:          santri.ID,
		NIS:               nis.Value(),
		PasswordGenerated: password,
	}, nil
}

var _ = time.Now

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
