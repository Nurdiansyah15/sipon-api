package rolepermission

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
)

type ListRolesUseCase struct {
	readModel port.RolePermissionQueryReadModel
}

func NewListRolesUseCase(readModel port.RolePermissionQueryReadModel) *ListRolesUseCase {
	return &ListRolesUseCase{readModel: readModel}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *ListRolesUseCase) Execute(ctx context.Context, req dto.ListRolesQuery) ([]dto.RoleResponse, dto.Meta, error) {
	if _, err := parseRoleType(req.RoleType, false); err != nil {
		return nil, dto.Meta{}, err
	}
	assignable, err := parseOptionalBool(req.Assignable)
	if err != nil {
		return nil, dto.Meta{}, err
	}
	result, meta, err := uc.readModel.ListRoles(ctx, port.RoleListReadQuery{
		RoleType:   strings.TrimSpace(req.RoleType),
		ScopeType:  strings.TrimSpace(req.ScopeType),
		Assignable: assignable,
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
	items := make([]dto.RoleResponse, 0, len(result))
	for _, item := range result {
		items = append(items, mapRoleReadItem(item))
	}
	return items, meta, nil
}
