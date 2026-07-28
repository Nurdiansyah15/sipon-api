package dto

import "time"

// ProfileResponse is the consolidated response for GET /api/v1/web/auth/profile.
// Merges UserMe + SessionData so the frontend gets profile, roles, and permissions
// in a single call.
type ProfileResponse struct {
	ID              string              `json:"id"`
	Username        string              `json:"username"`
	Fullname        *string             `json:"fullname"`
	Email           string              `json:"email"`
	IsEmailVerified bool                `json:"is_email_verified"`
	Phone           *string             `json:"phone,omitempty"`
	IsPhoneVerified bool                `json:"is_phone_verified"`
	Status          string              `json:"status"`
	HasPassword     bool                `json:"has_password"`
	CreatedAt       time.Time           `json:"created_at"`
	AvatarURL       *string             `json:"avatar_url,omitempty"`
	Roles           []SessionRole       `json:"roles"`
	Permissions     []SessionPermission `json:"permissions"`
}
