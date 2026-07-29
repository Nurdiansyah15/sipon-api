package entity

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
)

type SantriRequest struct {
	ID         string
	UserID     string
	NIS        *string
	Status     constant.SantriRequestStatus
	Notes      *string
	ReviewedBy *string
	ReviewedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func NewSantriRequest(id, userID string) (*SantriRequest, error) {
	if id == "" || userID == "" {
		return nil, domainerr.New(constant.CodeSantriRequestPersistenceFailed)
	}
	return &SantriRequest{
		ID:        id,
		UserID:    userID,
		Status:    constant.SantriRequestStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *SantriRequest) Approve(reviewerID, nis string) error {
	if r.Status != constant.SantriRequestStatusPending {
		return domainerr.New(constant.CodeSantriRequestInvalidStatus)
	}
	r.Status = constant.SantriRequestStatusApproved
	r.NIS = &nis
	now := time.Now()
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}

func (r *SantriRequest) Reject(reviewerID string, notes *string) error {
	if r.Status != constant.SantriRequestStatusPending {
		return domainerr.New(constant.CodeSantriRequestInvalidStatus)
	}
	r.Status = constant.SantriRequestStatusRejected
	r.Notes = notes
	now := time.Now()
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	r.UpdatedAt = now
	return nil
}
