package dto

import "time"

// ── Request DTOs ──────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=30"`
	Email    string  `json:"email"    binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Phone    *string `json:"phone,omitempty"`
	Fullname *string `json:"fullname,omitempty"`
	// DeviceID — lihat catatan di LoginRequest.DeviceID.
	DeviceID string `json:"device_id,omitempty"`
}

type LoginRequest struct {
	// Identifier bisa email, no telepon, atau username
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password"   binding:"required"`
	// DeviceID opsional — kalau dikirim, ditanam ke access & refresh token
	// supaya "logout device lain" bisa benar-benar mencabut akses device itu.
	DeviceID string `json:"device_id,omitempty"`
}

type RequestIdentityOTPRequest struct {
	Identifier string `json:"identifier" binding:"required"`
}

type VerifyIdentityOTPRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	OTP        string `json:"otp"        binding:"required,len=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordLocalRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// SetPasswordLocalRequest — set password lokal pertama kali untuk akun yang
// belum punya password (dibuat lewat NewLocalCredentialWithoutPassword).
// Tidak butuh current_password karena memang belum ada password untuk
// diverifikasi; ditolak (409) kalau akun ternyata sudah punya password lokal
// aktif (lihat User.SetLocalPassword).
type SetPasswordLocalRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Token    string `json:"token"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// RegisterResponse menyertakan token yang sudah diterbitkan (embed LoginResponse)
// supaya client bisa langsung login setelah register, tanpa panggilan /login terpisah.
type RegisterResponse struct {
	UserID string `json:"user_id"`
	LoginResponse
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         UserMe `json:"user"`
}

type UserMe struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	Fullname        *string   `json:"fullname"`
	Phone           *string   `json:"phone,omitempty"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	// HasPassword — false untuk akun yang belum pernah set password lokal.
	HasPassword bool `json:"has_password"`
}

type RequestIdentityOTPResponse struct {
	Message string `json:"message"`
}

type VerifyIdentityOTPResponse struct {
	Message string `json:"message"`
}

type ChangePasswordLocalResponse struct {
	Message string `json:"message"`
}

type SetPasswordLocalResponse struct {
	Message string `json:"message"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ── Change Identity (Email / Phone) ──────────────────────────────────────────

type RequestChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type RequestChangePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
}

type ConfirmChangeEmailRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}

type ConfirmChangePhoneRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}

type ChangeIdentityResponse struct {
	Message string `json:"message"`
}
