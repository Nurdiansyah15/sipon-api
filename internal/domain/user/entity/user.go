package entity

import (
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/valueobject"
	"time"
)

// ── User (Aggregate Root) ─────────────────────────────────────────────────────

type User struct {
	ID                  string
	Username            valueobject.Username
	Fullname            *string
	Email               valueobject.Email
	PhoneNumber         *valueobject.PhoneNumber
	Status              constant.UserStatus
	Credentials         []*Credential
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         *time.Time
	DeletedAt           *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

func NewUser(id string, username valueobject.Username, fullname *string, email valueobject.Email, phoneNumber *valueobject.PhoneNumber) (*User, error) {
	if id == "" {
		return nil, domainerr.New(constant.CodeUserIDRequired)
	}
	if email == (valueobject.Email{}) {
		return nil, domainerr.New(constant.CodeUserEmailRequired)
	}
	if phoneNumber != nil && *phoneNumber == (valueobject.PhoneNumber{}) {
		return nil, domainerr.New(constant.CodeUserPhoneNumberInvalid)
	}
	return &User{
		ID:          id,
		Username:    username,
		Fullname:    fullname,
		Email:       email,
		Status:      constant.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		PhoneNumber: phoneNumber,
	}, nil
}

// AddCredential menambah metode login baru ke user ini
func (u *User) AddCredential(c *Credential) {
	u.Credentials = append(u.Credentials, c)
	u.UpdatedAt = time.Now()
}

func (u *User) FindLoginIdentity(kind constant.LoginIdentifierKind, value string) *LoginIdentity {
	for _, cred := range u.Credentials {
		if identity := cred.FindLoginIdentity(kind, value); identity != nil {
			return identity
		}
	}
	return nil
}

// FindLoginIdentityByKind mencari identity berdasarkan jenis saja (tanpa value).
func (u *User) FindLoginIdentityByKind(kind constant.LoginIdentifierKind) *LoginIdentity {
	for _, cred := range u.Credentials {
		if identity := cred.FindLoginIdentityByKind(kind); identity != nil {
			return identity
		}
	}
	return nil
}

func (u *User) FindCredential(typ constant.CredentialType) *Credential {
	for _, c := range u.Credentials {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

// EnsureCanLogin memvalidasi apakah user boleh login (business rule)
func (u *User) EnsureCanLogin() error {
	if u.DeletedAt != nil {
		return domainerr.New(constant.CodeUserDeleted)
	}
	if u.Status == constant.StatusBanned {
		return domainerr.New(constant.CodeUserBanned)
	}
	if len(u.Credentials) == 0 {
		return domainerr.New(constant.CodeUserNoCredential)
	}
	return nil
}

// Activate mengaktifkan akun setelah verifikasi OTP
func (u *User) Activate() error {
	u.Status = constant.StatusActive
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) MarkLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = time.Now()
}

func (u *User) IsLockedOut() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

func (u *User) EnsureNotLockedOut() error {
	if u.IsLockedOut() {
		return domainerr.New(constant.CodeUserLockedOut)
	}
	return nil
}

func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= constant.MaxLoginAttempts {
		t := time.Now().Add(constant.LockoutDuration)
		u.LockedUntil = &t
	}
	u.UpdatedAt = time.Now()
}

func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
}

// HasLocalPassword mengecek apakah user punya credential LOCAL dengan password
// aktif. Dipakai sebagai guard sebelum operasi set-password pertama kali.
func (u *User) HasLocalPassword() bool {
	local := u.FindCredential(constant.CredentialTypeLocal)
	return local != nil && local.DeletedAt == nil && local.SecretHash != nil
}

// SetLocalPassword menetapkan password lokal untuk pertama kali pada credential
// LOCAL yang sudah ada (dibuat tanpa password saat registrasi, lihat
// NewLocalCredentialWithoutPassword). Ditolak kalau user sudah punya password
// lokal aktif — kasus itu harus lewat alur ganti password biasa (verifikasi
// password lama), bukan endpoint set-password ini.
func (u *User) SetLocalPassword(hashed valueobject.HashedPassword) error {
	if u.HasLocalPassword() {
		return domainerr.New(constant.CodeUserAlreadyHasLocalPassword)
	}
	local := u.FindCredential(constant.CredentialTypeLocal)
	if local == nil || local.DeletedAt != nil {
		return domainerr.New(constant.CodeUserNoLocalCredential)
	}

	now := time.Now()
	local.SecretHash = &hashed
	local.LastChangedAt = &now
	local.UpdatedAt = now
	u.UpdatedAt = now
	return nil
}

// ChangeUsername mengganti username user dan menyinkronkan LoginIdentity bertipe USERNAME.
// Invariant: semua LoginIdentity KIND=USERNAME yang aktif harus ikut diperbarui agar
// login via username tetap bekerja setelah perubahan.
func (u *User) ChangeUsername(newUsername valueobject.Username) {
	now := time.Now()
	u.Username = newUsername
	u.UpdatedAt = now
	for _, cred := range u.Credentials {
		for _, identity := range cred.LoginIdentities {
			if identity.Kind == constant.LoginIdentifierUsername && identity.DeletedAt == nil {
				identity.Value = newUsername.Value()
				identity.UpdatedAt = now
			}
		}
	}
}
