package usermanagement

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	userconstant "sipon-api/internal/domain/user/constant"
	userentity "sipon-api/internal/domain/user/entity"
	"sipon-api/internal/domain/user/valueobject"

	"github.com/google/uuid"
)

// mapUserDomainError menerjemahkan domain error dari UserRepo / User entity ke
// AppError. Pakai domain error code sebagai message string supaya kode spesifik
// dapat dibaca oleh client (konsisten dengan rolepermission helpers).
func mapUserDomainError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case userconstant.CodeUserNotFound, userconstant.CodeLoginIdentityNotFound:
			return apperror.NotFound(string(de.Code), err)
		case userconstant.CodeUserDuplicate:
			return apperror.Conflict(string(de.Code), err)
		case userconstant.CodeUserAlreadyBanned, userconstant.CodeUserAlreadyActive:
			return apperror.Conflict(string(de.Code), err)
		case userconstant.CodeUserQueryFailed, userconstant.CodeUserPersistenceFailed:
			return apperror.Internal(string(apperror.CodeInternal), err)
		default:
			return apperror.Unprocessable(string(de.Code), nil, err)
		}
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

// mapUserValidationError memetakan validasi valueobject (email/username/phone)
// ke 422.
func mapUserValidationError(err error) error {
	var de *domainerr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case userconstant.CodeInvalidEmailFormat,
			userconstant.CodeInvalidUsernameFormat,
			userconstant.CodeInvalidPhoneNumberFormat,
			userconstant.CodeUserIDRequired,
			userconstant.CodeUserEmailRequired,
			userconstant.CodeUserPhoneNumberInvalid:
			return apperror.Unprocessable(string(de.Code), nil, err)
		}
	}
	return apperror.Internal(string(apperror.CodeInternal), err)
}

// generateTemporaryPassword menghasilkan password acak yang memenuhi aturan
// valueobject.NewPlainPassword (min 8, ≥1 uppercase, ≥1 digit). Pakai
// crypto/rand agar tidak bisa diprediksi.
func generateTemporaryPassword() (string, error) {
	const (
		upper = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		digit = "23456789"
		rest  = "abcdefghijkmnpqrstuvwxyz"
	)
	pool := upper + digit + rest
	const length = 12

	buf := make([]byte, length)
	for i := range buf {
		b, err := randByte()
		if err != nil {
			return "", err
		}
		buf[i] = pool[int(b)%len(pool)]
	}
	if !hasUpper(buf) {
		b, err := randByte()
		if err != nil {
			return "", err
		}
		buf[int(b)%length] = upper[int(b)%len(upper)]
	}
	if !hasDigit(buf) {
		b, err := randByte()
		if err != nil {
			return "", err
		}
		buf[int(b)%length] = digit[int(b)%len(digit)]
	}
	return string(buf), nil
}

func randByte() (byte, error) {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func hasUpper(b []byte) bool {
	for _, c := range b {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func hasDigit(b []byte) bool {
	for _, c := range b {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// buildUserManagementResponse memetakan user entity menjadi DTO. Jika
// includeRoles=true, role summary aktif di-load dari read model.
func buildUserManagementResponse(ctx context.Context, readModel port.UserQueryReadModel, user *userentity.User, includeRoles bool) (*dto.UserManagementResponse, error) {
	var phone *string
	if user.PhoneNumber != nil {
		v := user.PhoneNumber.Value()
		phone = &v
	}
	resp := &dto.UserManagementResponse{
		ID:          user.ID,
		Username:    user.Username.Value(),
		Fullname:    user.Fullname,
		Email:       user.Email.Value(),
		Phone:       phone,
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
	if includeRoles && readModel != nil {
		roles, err := readModel.ListActiveRoleSummariesByUserID(ctx, user.ID)
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		resp.Roles = make([]dto.UserRoleSummaryResponse, 0, len(roles))
		for _, r := range roles {
			resp.Roles = append(resp.Roles, dto.UserRoleSummaryResponse{
				ID:        r.ID,
				RoleID:    r.RoleID,
				RoleName:  r.RoleName,
				ScopeType: r.ScopeType,
				ScopeID:   r.ScopeID,
				IsActive:  r.IsActive,
			})
		}
	}
	return resp, nil
}

// buildCredentialWithIdentities membentuk Credential LOCAL + identity
// EMAIL/USERNAME/PHONE aktif, mirip register tetapi tanpa jalur OTP/verifikasi
// (admin-create melewati verifikasi identity — lihat plan §5 create_user).
func buildCredentialWithIdentities(userID string, hashed valueobject.HashedPassword, usernameValue, emailValue string, phone *string) (*userentity.Credential, error) {
	credID := uuid.NewString()
	cred := userentity.NewLocalCredential(credID, userID, hashed, true)

	emailIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credID, userconstant.LoginIdentifierEmail, emailValue, true, nil)
	if err != nil {
		return nil, err
	}
	cred.AddLoginIdentity(emailIdentity)

	usernameIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credID, userconstant.LoginIdentifierUsername, usernameValue, true, nil)
	if err != nil {
		return nil, err
	}
	cred.AddLoginIdentity(usernameIdentity)

	if phone != nil && strings.TrimSpace(*phone) != "" {
		phoneIdentity, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credID, userconstant.LoginIdentifierPhone, *phone, false, nil)
		if err != nil {
			return nil, err
		}
		cred.AddLoginIdentity(phoneIdentity)
	}
	return cred, nil
}