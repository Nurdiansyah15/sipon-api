package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"
)

type AssignRoleInput struct {
	UserID     string
	RoleName   roleconstant.RoleName
	ScopeType  roleconstant.ScopeType
	ScopeID    *string
	AssignedBy string
	ExpiredAt  *time.Time
}

type UserRoleAssignmentService struct {
	roleRepo     rolerepo.RoleRepository
	userRoleRepo rolerepo.UserRoleRepository
}

func NewUserRoleAssignmentService(
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
) *UserRoleAssignmentService {
	return &UserRoleAssignmentService{roleRepo: roleRepo, userRoleRepo: userRoleRepo}
}

func (s *UserRoleAssignmentService) AssignByRoleName(ctx context.Context, input AssignRoleInput) (*roleentity.UserRole, error) {
	role, err := s.roleRepo.FindByName(ctx, input.RoleName)
	if err != nil {
		return nil, err
	}
	if err := role.EnsureAssignable(); err != nil {
		return nil, err
	}
	if err := role.EnsureAssignmentScopeMatch(input.ScopeType); err != nil {
		return nil, err
	}

	userID := strings.TrimSpace(input.UserID)
	assignedBy := strings.TrimSpace(input.AssignedBy)
	if userID == "" {
		return nil, domainerr.New(roleconstant.CodeRoleUserIDRequired)
	}
	if assignedBy == "" {
		return nil, domainerr.New(roleconstant.CodeRoleUserIDRequired)
	}

	activeAssignments, err := s.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, active := range activeAssignments {
		if active == nil {
			continue
		}
		if active.RoleID != role.ID || active.ScopeType != input.ScopeType {
			continue
		}
		if sameScopeID(active.ScopeID, input.ScopeID) {
			return active, nil
		}
	}

	assignment, err := roleentity.NewUserRole(
		uuid.NewString(),
		userID,
		role.ID,
		input.ScopeType,
		input.ScopeID,
		assignedBy,
		input.ExpiredAt,
	)
	if err != nil {
		return nil, err
	}
	if err := s.userRoleRepo.Save(ctx, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func sameScopeID(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return strings.TrimSpace(*a) == strings.TrimSpace(*b)
}
