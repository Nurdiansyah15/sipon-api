package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/valueobject"
	"sipon-api/internal/domain/verification/constant"
	"time"
)

// VerificationCode adalah aggregate root untuk alur OTP verifikasi.
type VerificationCode struct {
	ID               string
	UserID           string
	Code             valueobject.OTPCode
	Purpose          constant.CodePurpose
	ExpiresAt        time.Time
	UsedAt           *time.Time
	CreatedAt        time.Time
	NewIdentityValue *string // diisi untuk alur change-email / change-phone
}

func NewVerificationCode(id, userID, rawCode string, purpose constant.CodePurpose, ttl time.Duration) (*VerificationCode, error) {
	code, err := valueobject.NewOTPCode(rawCode)
	if err != nil {
		return nil, err
	}
	return &VerificationCode{
		ID:        id,
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}, nil
}

// Verify menerapkan business rule OTP di dalam aggregate.
func (v *VerificationCode) Verify(inputCode string) error {
	if time.Now().After(v.ExpiresAt) {
		return domainerr.New(constant.CodeOTPExpired)
	}
	if v.UsedAt != nil {
		return domainerr.New(constant.CodeOTPUsed)
	}
	if !v.Code.Match(inputCode) {
		return domainerr.New(constant.CodeOTPInvalid)
	}
	now := time.Now()
	v.UsedAt = &now
	return nil
}

func (v *VerificationCode) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// SetNewIdentityValue menyimpan nilai identitas baru yang ingin diubah user.
// Dipanggil setelah konstruksi, hanya untuk purpose ChangeEmail / ChangePhone.
func (v *VerificationCode) SetNewIdentityValue(val string) {
	v.NewIdentityValue = &val
}
