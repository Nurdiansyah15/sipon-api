package web

import (
	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	rolepermissionusecase "sipon-api/internal/app/usecase/rolepermission"
	"sipon-api/internal/interfaces/http/httperror"
	"sipon-api/internal/interfaces/http/respond"

	"github.com/gin-gonic/gin"
)

type RolePermissionHandler struct {
	useCases *rolepermissionusecase.UseCases
}

func NewRolePermissionHandler(useCases *rolepermissionusecase.UseCases) *RolePermissionHandler {
	return &RolePermissionHandler{useCases: useCases}
}

// ListRoles godoc
// @Summary Get all roles
// @Description Mengambil daftar roles dengan filter opsional.
// @Tags Web/RolePermission
// @Produce json
// @Param query query dto.ListRolesQuery false "Query parameters"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles [get]
func (h *RolePermissionHandler) ListRoles(c *gin.Context) {
	var req dto.ListRolesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	data, meta, err := h.useCases.ListRoles.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "roles fetched", data, meta)
}

// GetRole godoc
// @Summary Get role by ID
// @Description Mengambil detail role berdasarkan role_id.
// @Tags Web/RolePermission
// @Produce json
// @Param role_id path string true "Role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id} [get]
func (h *RolePermissionHandler) GetRole(c *gin.Context) {
	resp, err := h.useCases.GetRole.Execute(c.Request.Context(), c.Param("role_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role fetched", resp)
}

// CreateRole godoc
// @Summary Create role
// @Description Membuat master role baru.
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param request body dto.CreateRoleRequest true "Create role payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles [post]
func (h *RolePermissionHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.CreateRole.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "role created", resp)
}

// UpdateRole godoc
// @Summary Update role
// @Description Memperbarui role metadata.
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Param request body dto.UpdateRoleRequest true "Update role payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id} [put]
func (h *RolePermissionHandler) UpdateRole(c *gin.Context) {
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.UpdateRole.Execute(c.Request.Context(), c.Param("role_id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role updated", resp)
}

// ListPermissionKeys godoc
// @Summary Get permission key catalog
// @Description Mengambil daftar seluruh permission key yang dikenal sistem (didefinisikan di kode) — dipakai untuk memilih permission saat assign ke custom role.
// @Tags Web/RolePermission
// @Produce json
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/permission-keys [get]
func (h *RolePermissionHandler) ListPermissionKeys(c *gin.Context) {
	resp := h.useCases.ListPermissionKeys.Execute(c.Request.Context())
	respond.OK(c, "permission keys fetched", resp)
}

// AssignRolePermission godoc
// @Summary Assign permission to a custom role
// @Description Menambahkan permission ke role custom. Ditolak (409) untuk role system — permission role system fixed di kode.
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Param request body dto.AssignRolePermissionRequest true "Assign role permission payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id}/permissions [post]
func (h *RolePermissionHandler) AssignRolePermission(c *gin.Context) {
	var req dto.AssignRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	actorUserID, err := currentUserID(c)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.AssignRolePermission.Execute(c.Request.Context(), actorUserID, c.Param("role_id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "permission assigned to role", resp)
}

// RevokeRolePermission godoc
// @Summary Revoke permission from a custom role
// @Description Menghapus permission dari role custom. Ditolak (409) untuk role system.
// @Tags Web/RolePermission
// @Produce json
// @Param role_id path string true "Role ID"
// @Param permission_key path string true "Permission key"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id}/permissions/{permission_key} [delete]
func (h *RolePermissionHandler) RevokeRolePermission(c *gin.Context) {
	resp, err := h.useCases.RevokeRolePermission.Execute(c.Request.Context(), c.Param("role_id"), c.Param("permission_key"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "permission revoked from role", resp)
}

// ListUserRoles godoc
// @Summary Get all user roles
// @Description Mengambil daftar user role dengan filter opsional.
// @Tags Web/RolePermission
// @Produce json
// @Param query query dto.ListUserRolesQuery false "Query parameters"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles [get]
func (h *RolePermissionHandler) ListUserRoles(c *gin.Context) {
	var req dto.ListUserRolesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	data, meta, err := h.useCases.ListUserRoles.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "user roles fetched", data, meta)
}

// GetUserRole godoc
// @Summary Get user role by ID
// @Description Mengambil detail user role berdasarkan user_role_id.
// @Tags Web/RolePermission
// @Produce json
// @Param user_role_id path string true "User role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles/{user_role_id} [get]
func (h *RolePermissionHandler) GetUserRole(c *gin.Context) {
	resp, err := h.useCases.GetUserRole.Execute(c.Request.Context(), c.Param("user_role_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user role fetched", resp)
}

// AssignUserRole godoc
// @Summary Assign user role
// @Description Menambahkan assignment role pada user.
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param request body dto.AssignUserRoleRequest true "Assign user role payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles [post]
func (h *RolePermissionHandler) AssignUserRole(c *gin.Context) {
	var req dto.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	actorUserID, err := currentUserID(c)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.AssignUserRole.Execute(c.Request.Context(), actorUserID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "user role assigned", resp)
}

// UpdateUserRole godoc
// @Summary Update user role
// @Description Memperbarui expire date pada assignment user role.
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param user_role_id path string true "User role ID"
// @Param request body dto.UpdateUserRoleRequest true "Update user role payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles/{user_role_id} [put]
func (h *RolePermissionHandler) UpdateUserRole(c *gin.Context) {
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.UpdateUserRole.Execute(c.Request.Context(), c.Param("user_role_id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user role updated", resp)
}

// DeactivateUserRole godoc
// @Summary Deactivate user role
// @Description Menonaktifkan assignment user role.
// @Tags Web/RolePermission
// @Produce json
// @Param user_role_id path string true "User role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles/{user_role_id}/deactivate [post]
func (h *RolePermissionHandler) DeactivateUserRole(c *gin.Context) {
	resp, err := h.useCases.DeactivateUserRole.Execute(c.Request.Context(), c.Param("user_role_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user role deactivated", resp)
}

// ReactivateUserRole godoc
// @Summary Reactivate user role
// @Description Mengaktifkan kembali assignment user role.
// @Tags Web/RolePermission
// @Produce json
// @Param user_role_id path string true "User role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles/{user_role_id}/reactivate [post]
func (h *RolePermissionHandler) ReactivateUserRole(c *gin.Context) {
	resp, err := h.useCases.ReactivateUserRole.Execute(c.Request.Context(), c.Param("user_role_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user role reactivated", resp)
}

// DeleteUserRole godoc
// @Summary Delete user role
// @Description Menghapus assignment user role.
// @Tags Web/RolePermission
// @Produce json
// @Param user_role_id path string true "User role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/user-roles/{user_role_id} [delete]
func (h *RolePermissionHandler) DeleteUserRole(c *gin.Context) {
	if err := h.useCases.DeleteUserRole.Execute(c.Request.Context(), c.Param("user_role_id")); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User role deleted successfully", gin.H{"message": "User role deleted successfully"})
}

// ── Role Scopes ─────────────────────────────────────────────────────────────

// ListRoleScopes godoc
// @Summary List scopes for a role
// @Description Mengambil daftar scope yang di-assign ke role tertentu.
// @Tags Web/RolePermission
// @Produce json
// @Param role_id path string true "Role ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id}/scopes [get]
func (h *RolePermissionHandler) ListRoleScopes(c *gin.Context) {
	data, err := h.useCases.ListRoleScopes.Execute(c.Request.Context(), c.Param("role_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role scopes fetched", data)
}

// AssignRoleScope godoc
// @Summary Assign scope to role
// @Description Menambahkan scope baru ke role (hanya untuk custom role).
// @Tags Web/RolePermission
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Param request body dto.AssignRoleScopeRequest true "Scope payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id}/scopes [post]
func (h *RolePermissionHandler) AssignRoleScope(c *gin.Context) {
	var req dto.AssignRoleScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.AssignRoleScope.Execute(c.Request.Context(), c.Param("role_id"), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "role scope assigned", resp)
}

// RemoveRoleScope godoc
// @Summary Remove scope from role
// @Description Menghapus scope yang sudah di-assign ke role.
// @Tags Web/RolePermission
// @Produce json
// @Param role_id path string true "Role ID"
// @Param scope_id path string true "Scope ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/role-permission/roles/{role_id}/scopes/{scope_id} [delete]
func (h *RolePermissionHandler) RemoveRoleScope(c *gin.Context) {
	err := h.useCases.RemoveRoleScope.Execute(c.Request.Context(), c.Param("scope_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role scope removed", nil)
}

func currentUserID(c *gin.Context) (string, error) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		return "", apperror.Unauthorized("unauthorized")
	}
	userID, ok := userIDAny.(string)
	if !ok || userID == "" {
		return "", apperror.Unauthorized("unauthorized")
	}
	return userID, nil
}
