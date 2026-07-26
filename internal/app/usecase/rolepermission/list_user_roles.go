package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
)

type ListUserRolesUseCase struct {
	readModel port.RolePermissionQueryReadModel
}

func NewListUserRolesUseCase(readModel port.RolePermissionQueryReadModel) *ListUserRolesUseCase {
	return &ListUserRolesUseCase{readModel: readModel}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *ListUserRolesUseCase) Execute(ctx context.Context, req dto.ListUserRolesQuery) ([]dto.UserRoleResponse, dto.Meta, error) {
	if _, err := parseScopeType(req.ScopeType, false); err != nil {
		return nil, dto.Meta{}, err
	}
	isActive, err := parseOptionalBool(req.IsActive)
	if err != nil {
		return nil, dto.Meta{}, err
	}
	result, meta, err := uc.readModel.ListUserRoles(ctx, port.UserRoleListReadQuery{
		UserID:    strings.TrimSpace(req.UserID),
		RoleID:    strings.TrimSpace(req.RoleID),
		ScopeType: strings.TrimSpace(req.ScopeType),
		ScopeID:   strings.TrimSpace(req.ScopeID),
		IsActive:  isActive,
		PaginationParams: dto.PaginationParams{
			Page:     req.Page,
			Limit:    req.Limit,
			SortBy:   req.SortBy,
			SortType: req.SortType,
		},
	})
	if err != nil {
		return nil, dto.Meta{}, apperror.Internal(string(apperror.CodeInternal), err)
	}
	items := make([]dto.UserRoleResponse, 0, len(result))
	for _, item := range result {
		items = append(items, mapUserRoleReadItem(item))
	}
	return items, meta, nil
}
