package web

import (
	"sipon-api/internal/app/dto"
	usermanagementusecase "sipon-api/internal/app/usecase/usermanagement"
	"sipon-api/internal/interfaces/http/httperror"
	"sipon-api/internal/interfaces/http/respond"

	"github.com/gin-gonic/gin"
)

type UserManagementHandler struct {
	useCases *usermanagementusecase.UseCases
}

func NewUserManagementHandler(useCases *usermanagementusecase.UseCases) *UserManagementHandler {
	return &UserManagementHandler{useCases: useCases}
}

// ListUsers godoc
// @Summary List users (admin)
// @Description Mengambil daftar user (admin) dengan filter status/role_id/search dan pagination.
// @Tags Web/UserManagement
// @Produce json
// @Param query query dto.ListUsersQuery false "Query parameters"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users [get]
func (h *UserManagementHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	data, meta, err := h.useCases.ListUsers.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "users fetched", data, meta)
}

// GetUser godoc
// @Summary Get user by ID (admin)
// @Description Mengambil detail user beserta role assignment aktif-nya.
// @Tags Web/UserManagement
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users/{user_id} [get]
func (h *UserManagementHandler) GetUser(c *gin.Context) {
	resp, err := h.useCases.GetUser.Execute(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user fetched", resp)
}

// CreateUser godoc
// @Summary Create user (admin)
// @Description Membuat user baru dengan password auto-generated. Password hanya ditampilkan satu kali di response.
// @Tags Web/UserManagement
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "Create user payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users [post]
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.useCases.CreateUser.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "user created", resp)
}

// ResetUserPassword godoc
// @Summary Reset user password (admin)
// @Description Menyetel ulang kata sandi user lain. Password baru ditampilkan satu kali di response.
// @Tags Web/UserManagement
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users/{user_id}/reset-password [post]
func (h *UserManagementHandler) ResetUserPassword(c *gin.Context) {
	resp, err := h.useCases.ResetUserPassword.Execute(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user password reset", resp)
}

// DeactivateUser godoc
// @Summary Deactivate user (admin)
// @Description Menonaktifkan akun user (status BANNED).
// @Tags Web/UserManagement
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users/{user_id}/deactivate [post]
func (h *UserManagementHandler) DeactivateUser(c *gin.Context) {
	resp, err := h.useCases.DeactivateUser.Execute(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user deactivated", resp)
}

// ReactivateUser godoc
// @Summary Reactivate user (admin)
// @Description Mengaktifkan kembali akun user yang sebelumnya di-deactivate.
// @Tags Web/UserManagement
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/users/{user_id}/reactivate [post]
func (h *UserManagementHandler) ReactivateUser(c *gin.Context) {
	resp, err := h.useCases.ReactivateUser.Execute(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user reactivated", resp)
}