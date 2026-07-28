package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	authUsecase "sipon-api/internal/app/usecase/auth"
	"sipon-api/internal/interfaces/http/httperror"
	"sipon-api/internal/interfaces/http/middleware"
	"sipon-api/internal/interfaces/http/respond"
)

type AuthHandler struct {
	register              *authUsecase.RegisterUseCase
	login                 *authUsecase.LoginUseCase
	refreshToken          *authUsecase.RefreshTokenUseCase
	changePassword        *authUsecase.ChangePasswordLocalUseCase
	setPassword           *authUsecase.SetPasswordLocalUseCase
	requestIdentityOTP    *authUsecase.RequestIdentityOTPUseCase
	verifyIdentityOTP     *authUsecase.VerifyIdentityOTPUseCase
	me                    *authUsecase.MeUseCase
	forgotPassword        *authUsecase.ForgotPasswordUseCase
	resetPassword         *authUsecase.ResetPasswordUseCase
	requestChangeIdentity *authUsecase.RequestChangeIdentityUseCase
	confirmChangeIdentity *authUsecase.ConfirmChangeIdentityUseCase
	getSession            *authUsecase.GetSessionUseCase
	logout                *authUsecase.LogoutUseCase
	getProfile            *authUsecase.GetProfileUseCase
	updateProfile         *authUsecase.UpdateProfileUseCase
	checkUsername         *authUsecase.CheckUsernameUseCase
	changeUsername        *authUsecase.ChangeUsernameUseCase
	avatarPresign         *authUsecase.AvatarPresignUseCase
	avatarConfirm         *authUsecase.AvatarConfirmUseCase
	avatarDelete          *authUsecase.AvatarDeleteUseCase
}

func NewAuthHandler(
	register *authUsecase.RegisterUseCase,
	login *authUsecase.LoginUseCase,
	refreshToken *authUsecase.RefreshTokenUseCase,
	changePassword *authUsecase.ChangePasswordLocalUseCase,
	setPassword *authUsecase.SetPasswordLocalUseCase,
	requestIdentityOTP *authUsecase.RequestIdentityOTPUseCase,
	verifyIdentityOTP *authUsecase.VerifyIdentityOTPUseCase,
	me *authUsecase.MeUseCase,
	forgotPassword *authUsecase.ForgotPasswordUseCase,
	resetPassword *authUsecase.ResetPasswordUseCase,
	requestChangeIdentity *authUsecase.RequestChangeIdentityUseCase,
	confirmChangeIdentity *authUsecase.ConfirmChangeIdentityUseCase,
	getSession *authUsecase.GetSessionUseCase,
	logout *authUsecase.LogoutUseCase,
	getProfile *authUsecase.GetProfileUseCase,
	updateProfile *authUsecase.UpdateProfileUseCase,
	checkUsername *authUsecase.CheckUsernameUseCase,
	changeUsername *authUsecase.ChangeUsernameUseCase,
	avatarPresign *authUsecase.AvatarPresignUseCase,
	avatarConfirm *authUsecase.AvatarConfirmUseCase,
	avatarDelete *authUsecase.AvatarDeleteUseCase,
) *AuthHandler {
	return &AuthHandler{
		register:              register,
		login:                 login,
		refreshToken:          refreshToken,
		changePassword:        changePassword,
		setPassword:           setPassword,
		requestIdentityOTP:    requestIdentityOTP,
		verifyIdentityOTP:     verifyIdentityOTP,
		me:                    me,
		forgotPassword:        forgotPassword,
		resetPassword:         resetPassword,
		requestChangeIdentity: requestChangeIdentity,
		confirmChangeIdentity: confirmChangeIdentity,
		getSession:            getSession,
		logout:                logout,
		getProfile:            getProfile,
		updateProfile:         updateProfile,
		checkUsername:         checkUsername,
		changeUsername:        changeUsername,
		avatarPresign:         avatarPresign,
		avatarConfirm:         avatarConfirm,
		avatarDelete:          avatarDelete,
	}
}

// Register godoc
// @Summary Register user
// @Description Registrasi user baru dengan email dan opsional no hp untuk verifikasi OTP.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register payload"
// @Success 201 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/register [post]
// POST /api/v1/web/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.register.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.Created(c, "register success", resp)
}

// Login godoc
// @Summary Login user
// @Description Login dengan email, no hp, atau username.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/login [post]
// POST /api/v1/web/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.login.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "login success", resp)
}

// RefreshToken godoc
// @Summary Refresh auth token
// @Description Membuat access token dan refresh token baru dari refresh token yang valid.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/refresh-token [post]
// POST /api/v1/web/auth/refresh-token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.refreshToken.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "refresh token success", resp)
}

// ChangePasswordLocal godoc
// @Summary Change local password
// @Description Mengubah password akun local (wajib kirim password lama).
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordLocalRequest true "Change password payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 403 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-password [post]
// POST /api/v1/web/auth/change-password
func (h *AuthHandler) ChangePasswordLocal(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.ChangePasswordLocalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.changePassword.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "change password success", resp)
}

// SetPasswordLocal godoc
// @Summary Set local password (first time)
// @Description Menetapkan password lokal pertama kali untuk akun yang belum pernah punya password. Tidak perlu password lama.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.SetPasswordLocalRequest true "Set password payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/set-password [post]
// POST /api/v1/web/auth/set-password
func (h *AuthHandler) SetPasswordLocal(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.SetPasswordLocalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.setPassword.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "set password success", resp)
}

// RequestIdentityOTP godoc
// @Summary Request OTP verification
// @Description Mengirim OTP verifikasi ke email atau no hp berdasarkan identifier.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.RequestIdentityOTPRequest true "Request OTP payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/request-otp [post]
// POST /api/v1/web/auth/request-otp
func (h *AuthHandler) RequestIdentityOTP(c *gin.Context) {
	var req dto.RequestIdentityOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.requestIdentityOTP.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "request otp success", resp)
}

// VerifyIdentityOTP godoc
// @Summary Verify identity OTP
// @Description Verifikasi OTP email atau no hp dan ubah status identity menjadi verified.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyIdentityOTPRequest true "Verify OTP payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/verify-otp [post]
// POST /api/v1/web/auth/verify-otp
func (h *AuthHandler) VerifyIdentityOTP(c *gin.Context) {
	var req dto.VerifyIdentityOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.verifyIdentityOTP.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "verify otp success", resp)
}

// Me godoc
// @Summary Current user session
// @Description Mengembalikan payload user dari JWT yang sedang aktif.
// @Tags Web/Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Router /api/v1/web/auth/me [get]
// GET /api/v1/web/auth/me (protected)
func (h *AuthHandler) Me(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	resp, err := h.me.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "me success", resp)
}

// Profile godoc
// @Summary Get user profile with roles & permissions
// @Description Mengembalikan data profil user lengkap dengan roles dan permissions dalam satu response.
// @Tags Web/Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} respond.SuccessBody{data=dto.ProfileResponse}
// @Failure 401 {object} respond.ErrorBody
// @Router /api/v1/web/auth/profile [get]
// GET /api/v1/web/auth/profile (protected)
func (h *AuthHandler) Profile(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	if p == nil {
		httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
		return
	}

	resp, err := h.getProfile.Execute(c.Request.Context(), p.UserID, p)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "OK", resp)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Memperbarui profil user: fullname (selalu bisa), email & phone (hanya jika belum terverifikasi).
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update profile payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/profile [put]
// PUT /api/v1/web/auth/profile (protected)
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.updateProfile.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "update profile success", resp)
}

// CheckUsername godoc
// @Summary Check username availability
// @Description Mengecek apakah username tersedia (belum dipakai user lain).
// @Tags Web/Auth
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} respond.SuccessBody{data=dto.CheckUsernameResponse}
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/check-username [get]
// GET /api/v1/web/auth/check-username?username=xxx (protected)
func (h *AuthHandler) CheckUsername(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	username := c.Query("username")
	if username == "" {
		httperror.Handle(c, apperror.Unprocessable("parameter username wajib diisi", nil))
		return
	}

	resp, err := h.checkUsername.Execute(c.Request.Context(), userID, username)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "check username success", resp)
}

// ChangeUsername godoc
// @Summary Change username
// @Description Mengganti username (tidak perlu OTP). Langsung berubah jika tersedia.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ChangeUsernameRequest true "New username"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-username [post]
// POST /api/v1/web/auth/change-username (protected)
func (h *AuthHandler) ChangeUsername(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.ChangeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.changeUsername.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "change username success", resp)
}

// ── Avatar Upload ─────────────────────────────────────────────────────────

// AvatarPresign godoc
// @Summary Request presigned URL untuk upload avatar
// @Description Membuat presigned URL unggah avatar langsung ke object storage.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.AvatarPresignRequest true "Content-type gambar"
// @Success 200 {object} respond.SuccessBody{data=dto.AvatarPresignResponse}
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/profile/avatar/presign [post]
// POST /api/v1/web/auth/profile/avatar/presign (protected)
func (h *AuthHandler) AvatarPresign(c *gin.Context) {
	var req dto.AvatarPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.avatarPresign.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "avatar presign success", resp)
}

// AvatarConfirm godoc
// @Summary Konfirmasi upload avatar
// @Description Mengkonfirmasi bahwa avatar telah berhasil diupload ke object storage.
// @Tags Web/Auth
// @Produce json
// @Param key query string true "Object key dari response presign"
// @Success 200 {object} respond.SuccessBody{data=dto.AvatarConfirmResponse}
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/profile/avatar/confirm [post]
// POST /api/v1/web/auth/profile/avatar/confirm?key=... (protected)
func (h *AuthHandler) AvatarConfirm(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	key := c.Query("key")
	if key == "" {
		httperror.Handle(c, apperror.Unprocessable("parameter key wajib diisi", nil))
		return
	}

	resp, err := h.avatarConfirm.Execute(c.Request.Context(), userID, key)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "avatar confirm success", resp)
}

// AvatarDelete godoc
// @Summary Hapus avatar
// @Description Menghapus avatar user dari object storage dan database.
// @Tags Web/Auth
// @Produce json
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/profile/avatar [delete]
// DELETE /api/v1/web/auth/profile/avatar (protected)
func (h *AuthHandler) AvatarDelete(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	resp, err := h.avatarDelete.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "avatar delete success", resp)
}

// ForgotPassword godoc
// @Summary Request password reset OTP
// @Description Mengirim OTP reset password ke email. Selalu mengembalikan success untuk menghindari enumerasi user.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/password/forgot [post]
// POST /api/v1/web/auth/password/forgot
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.forgotPassword.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "forgot password success", resp)
}

// ResetPassword godoc
// @Summary Reset password with OTP
// @Description Reset password menggunakan OTP yang dikirim ke email.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset password payload"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/web/auth/password/reset [post]
// POST /api/v1/web/auth/password/reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.resetPassword.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "reset password success", resp)
}

// ── Change Identity (Email / Phone) ──────────────────────────────────────────

// RequestChangeEmail godoc
// @Summary Request OTP to change email
// @Description Mengirim OTP ke email baru. Email lama tidak berubah sampai dikonfirmasi.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.RequestChangeEmailRequest true "New email"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-email/request [post]
func (h *AuthHandler) RequestChangeEmail(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.RequestChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.requestChangeIdentity.Execute(c.Request.Context(), authUsecase.RequestChangeIdentityInput{
		UserID:   userID,
		Kind:     "EMAIL",
		NewValue: req.NewEmail,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, resp.Message, resp)
}

// ConfirmChangeEmail godoc
// @Summary Confirm email change via OTP
// @Description Verifikasi OTP untuk mengganti email. Email resmi berubah setelah endpoint ini berhasil.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ConfirmChangeEmailRequest true "OTP code"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-email/confirm [post]
func (h *AuthHandler) ConfirmChangeEmail(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.ConfirmChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.confirmChangeIdentity.Execute(c.Request.Context(), authUsecase.ConfirmChangeIdentityInput{
		UserID: userID,
		Kind:   "EMAIL",
		OTP:    req.OTP,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, resp.Message, resp)
}

// RequestChangePhone godoc
// @Summary Request OTP to change phone number
// @Description Mengirim OTP ke nomor baru. Nomor lama tidak berubah sampai dikonfirmasi.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.RequestChangePhoneRequest true "New phone number"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 409 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-phone/request [post]
func (h *AuthHandler) RequestChangePhone(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.RequestChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.requestChangeIdentity.Execute(c.Request.Context(), authUsecase.RequestChangeIdentityInput{
		UserID:   userID,
		Kind:     "PHONE",
		NewValue: req.NewPhone,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, resp.Message, resp)
}

// ConfirmChangePhone godoc
// @Summary Confirm phone number change via OTP
// @Description Verifikasi OTP untuk mengganti nomor telepon. Nomor resmi berubah setelah endpoint ini berhasil.
// @Tags Web/Auth
// @Accept json
// @Produce json
// @Param request body dto.ConfirmChangePhoneRequest true "OTP code"
// @Success 200 {object} respond.SuccessBody
// @Failure 400 {object} respond.ErrorBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 404 {object} respond.ErrorBody
// @Failure 422 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/web/auth/change-phone/confirm [post]
func (h *AuthHandler) ConfirmChangePhone(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}
	userID, ok := userIDAny.(string)
	if !ok {
		httperror.Handle(c, apperror.Unauthorized("unauthorized"))
		return
	}

	var req dto.ConfirmChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}

	resp, err := h.confirmChangeIdentity.Execute(c.Request.Context(), authUsecase.ConfirmChangeIdentityInput{
		UserID: userID,
		Kind:   "PHONE",
		OTP:    req.OTP,
	})
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, resp.Message, resp)
}

// GetSession godoc
// @Summary Session bootstrap
// @Description Mengembalikan data sesi lengkap setelah login: user, roles, permissions.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} respond.SuccessBody{data=dto.SessionData}
// @Failure 401 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Router /api/v1/auth/session [get]
// GET /api/v1/auth/session
func (h *AuthHandler) GetSession(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	if p == nil {
		respond.AbortWithError(c, http.StatusUnauthorized, string(apperror.CodeUnauthorized), "sesi tidak ditemukan")
		return
	}

	data, err := h.getSession.Execute(c.Request.Context(), p)
	if err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "OK", data)
}

// Logout godoc
// @Summary Logout device sekarang
// @Description Merevoke access token session yang sedang dipakai.
// @Tags Auth
// @Produce json
// @Success 200 {object} respond.SuccessBody
// @Failure 401 {object} respond.ErrorBody
// @Failure 500 {object} respond.ErrorBody
// @Security BearerAuth
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := c.GetString("session_id")
	if sessionID == "" {
		respond.AbortWithError(c, http.StatusUnauthorized, string(apperror.CodeUnauthorized), "sesi tidak ditemukan")
		return
	}

	if err := h.logout.Execute(c.Request.Context(), sessionID); err != nil {
		httperror.Handle(c, err)
		return
	}

	respond.OK(c, "Logout berhasil", nil)
}
