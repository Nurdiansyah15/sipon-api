package usermanagement

import (
	"context"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
)

type ListUsersUseCase struct {
	readModel port.UserQueryReadModel
}

func NewListUsersUseCase(readModel port.UserQueryReadModel) *ListUsersUseCase {
	return &ListUsersUseCase{readModel: readModel}
}

// Required — perm: manage_users
func (uc *ListUsersUseCase) Execute(ctx context.Context, req dto.ListUsersQuery) ([]dto.UserManagementResponse, dto.Meta, error) {
	result, meta, err := uc.readModel.ListUsers(ctx, port.UserListReadQuery{
		Status:  req.Status,
		RoleID:  req.RoleID,
		Search:  req.Search,
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
	items := make([]dto.UserManagementResponse, 0, len(result))
	userIDs := make([]string, 0, len(result))
	for _, item := range result {
		var phone *string
		if item.Phone != nil && *item.Phone != "" {
			p := *item.Phone
			phone = &p
		}
		items = append(items, dto.UserManagementResponse{
			ID:          item.ID,
			Username:    item.Username,
			Fullname:    item.Fullname,
			Email:       item.Email,
			Phone:       phone,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			LastLoginAt: item.LastLoginAt,
		})
		userIDs = append(userIDs, item.ID)
	}

	// Batch fetch active role summaries untuk semua user dalam satu query.
	if len(userIDs) > 0 {
		roleMap, err := uc.readModel.ListActiveRoleSummariesByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, dto.Meta{}, apperror.Internal(string(apperror.CodeInternal), err)
		}
		for i := range items {
			if roles, ok := roleMap[items[i].ID]; ok {
				items[i].Roles = make([]dto.UserRoleSummaryResponse, 0, len(roles))
				for _, r := range roles {
					items[i].Roles = append(items[i].Roles, dto.UserRoleSummaryResponse{
						ID:        r.ID,
						RoleID:    r.RoleID,
						RoleName:  r.RoleName,
						ScopeType: r.ScopeType,
						ScopeID:   r.ScopeID,
						IsActive:  r.IsActive,
					})
				}
			}
		}
	}

	return items, meta, nil
}