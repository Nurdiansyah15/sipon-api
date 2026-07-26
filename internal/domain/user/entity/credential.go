package entity

import (
	"sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/valueobject"
	"time"
)

// Credential adalah entity yang menyimpan satu metode autentikasi.
// Satu Credential bisa memiliki banyak LoginIdentity.
// Untuk provider "local": Secret = hashed password.
type Credential struct {
	ID              string
	UserID          string
	Type            constant.CredentialType
	LoginIdentities []*LoginIdentity
	SecretHash      *valueobject.HashedPassword
	LastChangedAt   *time.Time
	IsPrimary       bool
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
	DeletedAt       *time.Time
}

func NewLocalCredential(id, userID string, hashed valueobject.HashedPassword, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:            id,
		UserID:        userID,
		Type:          constant.CredentialTypeLocal,
		SecretHash:    &hashed,
		LastChangedAt: &now,
		IsPrimary:     isPrimary,
		UpdatedAt:     now,
	}
}

func NewLocalCredentialWithoutPassword(id, userID string, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:        id,
		UserID:    userID,
		Type:      constant.CredentialTypeLocal,
		IsPrimary: isPrimary,
		UpdatedAt: now,
	}
}

func (c *Credential) AddLoginIdentity(identity *LoginIdentity) {
	c.LoginIdentities = append(c.LoginIdentities, identity)
	c.UpdatedAt = time.Now()
}

func (c *Credential) FindLoginIdentity(kind constant.LoginIdentifierKind, value string) *LoginIdentity {
	for _, identity := range c.LoginIdentities {
		if identity.Kind == kind && identity.Value == value && identity.DeletedAt == nil {
			return identity
		}
	}
	return nil
}

// FindLoginIdentityByKind mencari identity berdasarkan jenis saja (tanpa value).
func (c *Credential) FindLoginIdentityByKind(kind constant.LoginIdentifierKind) *LoginIdentity {
	for _, identity := range c.LoginIdentities {
		if identity.Kind == kind && identity.DeletedAt == nil {
			return identity
		}
	}
	return nil
}
