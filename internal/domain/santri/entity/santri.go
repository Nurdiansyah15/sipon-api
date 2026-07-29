package entity

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/valueobject"
)

type Santri struct {
	ID     string
	UserID string
	NIS    *valueobject.NIS

	Nickname         *string
	Program          string
	Option           string
	Hobby            *string
	Purpose          *string
	MotivationEntry  *string
	POB              *string
	DOB              *time.Time
	Blood            *string

	Address      *string
	SubDistrict  *string
	District     *string
	Province     *string
	PostalCode   *string

	PreviousPondokName    *string
	PreviousPondokAddress *string
	PreviousPondokDiv     *string
	PreviousPondokTime    *string

	NIK   *string
	NoKK  *string
	NISN  *string
	NoKIP *string
	NoKKS *string
	NoPKH *string

	Workplace  *string
	Department *string

	HomeStatus *string

	Father          *string
	FatherPN        *string
	FatherNIK       *string
	FatherJob       *string
	FatherGraduate  *string
	FatherIncome    *string

	Mother          *string
	MotherPN        *string
	MotherNIK       *string
	MotherJob       *string
	MotherGraduate  *string
	MotherIncome    *string

	GuardianRelationship *string
	Guardian             *string
	GuardianPN           *string
	GuardianNIK          *string
	GuardianJob          *string
	GuardianGraduate     *string
	GuardianIncome       *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewSantri(id, userID string) (*Santri, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeSantriPersistenceFailed)
	}
	if userID == "" {
		return nil, domainerr.New(constant.CodeSantriPersistenceFailed)
	}
	return &Santri{
		ID:     id,
		UserID: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *Santri) Update() {
	s.UpdatedAt = time.Now()
}
