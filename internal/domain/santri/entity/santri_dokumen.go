package entity

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
)

type SantriDokumen struct {
	ID               string
	SantriID         string
	Kind             constant.DokumenKind
	Key              string
	Status           constant.DokumenStatus
	OriginalFilename *string
	MimeType         *string
	Size             *int64
	Notes            *string
	VerifiedBy       *string
	VerifiedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewSantriDokumen(id, santriID string, kind constant.DokumenKind, key string) (*SantriDokumen, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeSantriPersistenceFailed)
	}
	if santriID == "" {
		return nil, domainerr.New(constant.CodeSantriPersistenceFailed)
	}
	if !constant.ValidDokumenKinds[kind] {
		return nil, domainerr.New(constant.CodeDokumenInvalidKind)
	}
	if key == "" {
		return nil, domainerr.New(constant.CodeDokumenPersistenceFailed)
	}
	return &SantriDokumen{
		ID:        id,
		SantriID:  santriID,
		Kind:      kind,
		Key:       key,
		Status:    constant.DokumenStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (d *SantriDokumen) Verify(verifierID string) error {
	if d.Status == constant.DokumenStatusVerified {
		return nil
	}
	if d.Status == constant.DokumenStatusRejected {
		return domainerr.New(constant.CodeDokumenInvalidStatus)
	}
	d.Status = constant.DokumenStatusVerified
	now := time.Now()
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.UpdatedAt = now
	return nil
}

func (d *SantriDokumen) Reject(verifierID string, notes *string) error {
	if d.Status == constant.DokumenStatusRejected {
		return nil
	}
	d.Status = constant.DokumenStatusRejected
	now := time.Now()
	d.VerifiedBy = &verifierID
	d.VerifiedAt = &now
	d.Notes = notes
	d.UpdatedAt = now
	return nil
}

func (d *SantriDokumen) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}
