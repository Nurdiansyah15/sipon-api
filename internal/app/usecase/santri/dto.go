package santriUsecase

import "time"

// ── Santri Profile ────────────────────────────────────────────────────────────

type GetSantriResponse struct {
	ID    string  `json:"id"`
	UserID string `json:"user_id"`
	NIS   *string `json:"nis,omitempty"`

	Fullname  *string `json:"fullname,omitempty"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`

	Nickname        *string `json:"nickname,omitempty"`
	Program         string  `json:"program"`
	Option          string  `json:"option"`
	Hobby           *string `json:"hobby,omitempty"`
	Purpose         *string `json:"purpose,omitempty"`
	MotivationEntry *string `json:"motivation_entry,omitempty"`
	POB             *string `json:"pob,omitempty"`
	DOB             *string `json:"dob,omitempty"`
	Blood           *string `json:"blood,omitempty"`

	Address     *string `json:"address,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`

	PreviousPondokName    *string `json:"previous_pondok_name,omitempty"`
	PreviousPondokAddress *string `json:"previous_pondok_address,omitempty"`
	PreviousPondokDiv     *string `json:"previous_pondok_div,omitempty"`
	PreviousPondokTime    *string `json:"previous_pondok_time,omitempty"`

	NIK   *string `json:"nik,omitempty"`
	NoKK  *string `json:"no_kk,omitempty"`
	NISN  *string `json:"nisn,omitempty"`
	NoKIP *string `json:"no_kip,omitempty"`
	NoKKS *string `json:"no_kks,omitempty"`
	NoPKH *string `json:"no_pkh,omitempty"`

	Workplace  *string `json:"workplace,omitempty"`
	Department *string `json:"department,omitempty"`

	HomeStatus           *string `json:"home_status,omitempty"`
	Father               *string `json:"father,omitempty"`
	FatherPN             *string `json:"father_pn,omitempty"`
	FatherNIK            *string `json:"father_nik,omitempty"`
	FatherJob            *string `json:"father_job,omitempty"`
	FatherGraduate       *string `json:"father_graduate,omitempty"`
	FatherIncome         *string `json:"father_income,omitempty"`
	Mother               *string `json:"mother,omitempty"`
	MotherPN             *string `json:"mother_pn,omitempty"`
	MotherNIK            *string `json:"mother_nik,omitempty"`
	MotherJob            *string `json:"mother_job,omitempty"`
	MotherGraduate       *string `json:"mother_graduate,omitempty"`
	MotherIncome         *string `json:"mother_income,omitempty"`
	GuardianRelationship *string `json:"guardian_relationship,omitempty"`
	Guardian             *string `json:"guardian,omitempty"`
	GuardianPN           *string `json:"guardian_pn,omitempty"`
	GuardianNIK          *string `json:"guardian_nik,omitempty"`
	GuardianJob          *string `json:"guardian_job,omitempty"`
	GuardianGraduate     *string `json:"guardian_graduate,omitempty"`
	GuardianIncome       *string `json:"guardian_income,omitempty"`
}

type UpdateSantriRequest struct {
	Fullname *string `json:"fullname,omitempty"`

	Nickname        *string `json:"nickname,omitempty"`
	Hobby           *string `json:"hobby,omitempty"`
	Purpose         *string `json:"purpose,omitempty"`
	MotivationEntry *string `json:"motivation_entry,omitempty"`
	POB             *string `json:"pob,omitempty"`
	DOB             *string `json:"dob,omitempty"`
	Blood           *string `json:"blood,omitempty"`

	Address     *string `json:"address,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`

	PreviousPondokName    *string `json:"previous_pondok_name,omitempty"`
	PreviousPondokAddress *string `json:"previous_pondok_address,omitempty"`
	PreviousPondokDiv     *string `json:"previous_pondok_div,omitempty"`
	PreviousPondokTime    *string `json:"previous_pondok_time,omitempty"`

	NIK   *string `json:"nik,omitempty"`
	NoKK  *string `json:"no_kk,omitempty"`
	NISN  *string `json:"nisn,omitempty"`
	NoKIP *string `json:"no_kip,omitempty"`
	NoKKS *string `json:"no_kks,omitempty"`
	NoPKH *string `json:"no_pkh,omitempty"`

	Workplace  *string `json:"workplace,omitempty"`
	Department *string `json:"department,omitempty"`

	HomeStatus           *string `json:"home_status,omitempty"`
	Father               *string `json:"father,omitempty"`
	FatherPN             *string `json:"father_pn,omitempty"`
	FatherNIK            *string `json:"father_nik,omitempty"`
	FatherJob            *string `json:"father_job,omitempty"`
	FatherGraduate       *string `json:"father_graduate,omitempty"`
	FatherIncome         *string `json:"father_income,omitempty"`
	Mother               *string `json:"mother,omitempty"`
	MotherPN             *string `json:"mother_pn,omitempty"`
	MotherNIK            *string `json:"mother_nik,omitempty"`
	MotherJob            *string `json:"mother_job,omitempty"`
	MotherGraduate       *string `json:"mother_graduate,omitempty"`
	MotherIncome         *string `json:"mother_income,omitempty"`
	GuardianRelationship *string `json:"guardian_relationship,omitempty"`
	Guardian             *string `json:"guardian,omitempty"`
	GuardianPN           *string `json:"guardian_pn,omitempty"`
	GuardianNIK          *string `json:"guardian_nik,omitempty"`
	GuardianJob          *string `json:"guardian_job,omitempty"`
	GuardianGraduate     *string `json:"guardian_graduate,omitempty"`
	GuardianIncome       *string `json:"guardian_income,omitempty"`
}

type UpdateSantriResponse struct {
	Message string `json:"message"`
}

// ── Admin Create Santri ──────────────────────────────────────────────────────

type CreateSantriRequest struct {
	NIS string `json:"nis" binding:"required"`
}

type CreateSantriResponse struct {
	UserID            string `json:"user_id"`
	SantriID          string `json:"santri_id"`
	NIS               string `json:"nis"`
	PasswordGenerated string `json:"password_generated"`
}

// ── Admin List Santri ────────────────────────────────────────────────────────

type ListSantriQuery struct {
	NIS      *string `form:"nis"`
	Page     *int    `form:"page"`
	Limit    *int    `form:"limit"`
	SortBy   *string `form:"sort_by"`
	SortType *string `form:"sort_type"`
}

type ListSantriItem struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	NIS       *string `json:"nis,omitempty"`
	Fullname  *string `json:"fullname,omitempty"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Status    string  `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Santri Request ────────────────────────────────────────────────────────────

type RequestSantriResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ListSantriRequestsQuery struct {
	Status   *string `form:"status"`
	Page     *int    `form:"page"`
	Limit    *int    `form:"limit"`
	SortBy   *string `form:"sort_by"`
	SortType *string `form:"sort_type"`
}

type SantriRequestItem struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	Fullname   *string `json:"fullname,omitempty"`
	Email      string  `json:"email"`
	Status     string  `json:"status"`
	Notes      *string `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type ApproveSantriRequestRequest struct {
	NIS string `json:"nis" binding:"required"`
}

type RejectSantriRequestRequest struct {
	Notes *string `json:"notes,omitempty"`
}

// ── Dokumen Presign ───────────────────────────────────────────────────────────

type DokumenPresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Kind        string `json:"kind"         binding:"required"`
}

type DokumenPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}

type DokumenConfirmRequest struct {
	Kind             string  `json:"kind" binding:"required"`
	Key              string  `json:"key"  binding:"required"`
	OriginalFilename *string `json:"original_filename,omitempty"`
	MimeType         *string `json:"mime_type,omitempty"`
	Size             *int64  `json:"size,omitempty"`
}

type DokumenConfirmResponse struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Key       string    `json:"key"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type DokumenItem struct {
	ID               string     `json:"id"`
	Kind             string     `json:"kind"`
	Key              string     `json:"key"`
	Status           string     `json:"status"`
	OriginalFilename *string    `json:"original_filename,omitempty"`
	MimeType         *string    `json:"mime_type,omitempty"`
	Size             *int64     `json:"size,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	VerifiedBy       *string    `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type DokumenAccessResponse struct {
	AccessURL string `json:"access_url"`
	ExpiresIn int    `json:"expires_in"`
}

type VerifyDokumenRequest struct {
	Notes *string `json:"notes,omitempty"`
}
