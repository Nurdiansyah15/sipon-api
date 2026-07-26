package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/valueobject"
	"time"
)

type LoginIdentity struct {
	ID           string
	UserID       string
	CredentialID string
	Kind         constant.LoginIdentifierKind
	Value        string
	Status       constant.LoginIdentityStatus
	IsPrimary    bool
	VerifiedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewLoginIdentity(id, userID, credentialID string, kind constant.LoginIdentifierKind, rawValue string, isPrimary bool, verifiedAt *time.Time) (*LoginIdentity, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeUserIDRequired)
	}
	if userID == "" {
		return nil, domainerr.New(constant.CodeUserIDRequired)
	}
	if credentialID == "" {
		return nil, domainerr.New(constant.CodeUserIDRequired)
	}

	value, err := normalizeLoginIdentityValue(kind, rawValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	status := constant.LoginIdentityStatusUnverified
	if verifiedAt != nil {
		status = constant.LoginIdentityStatusVerified
	}

	return &LoginIdentity{
		ID:           id,
		UserID:       userID,
		CredentialID: credentialID,
		Kind:         kind,
		Value:        value,
		Status:       status,
		IsPrimary:    isPrimary,
		VerifiedAt:   verifiedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (i *LoginIdentity) MarkVerified() {
	now := time.Now()
	i.Status = constant.LoginIdentityStatusVerified
	i.VerifiedAt = &now
	i.UpdatedAt = now
}

func (i *LoginIdentity) IsVerified() bool {
	return i.Status == constant.LoginIdentityStatusVerified
}

func (i *LoginIdentity) EnsureVerified() error {
	if i.IsVerified() {
		return nil
	}
	return domainerr.New(constant.CodeLoginIdentityUnverified)
}

func normalizeLoginIdentityValue(kind constant.LoginIdentifierKind, rawValue string) (string, error) {
	switch kind {
	case constant.LoginIdentifierEmail:
		email, err := valueobject.NewEmail(rawValue)
		if err != nil {
			return "", err
		}
		return email.Value(), nil
	case constant.LoginIdentifierPhone:
		phone, err := valueobject.NewPhoneNumber(rawValue)
		if err != nil {
			return "", err
		}
		return phone.Value(), nil
	case constant.LoginIdentifierUsername:
		username, err := valueobject.NewUsername(rawValue)
		if err != nil {
			return "", err
		}
		return username.Value(), nil
	default:
		return "", domainerr.New(constant.CodeInvalidLoginIdentifier)
	}
}
