package constant

import domainerr "sipon-api/internal/domain/errors"

type DokumenKind string

const (
	DokumenKindSuratPernyataan DokumenKind = "surat_pernyataan"
	DokumenKindKTP             DokumenKind = "ktp"
	DokumenKindKK              DokumenKind = "kk"
	DokumenKindMutasi          DokumenKind = "mutasi"
	DokumenKindPembayaran      DokumenKind = "pembayaran"
)

var ValidDokumenKinds = map[DokumenKind]bool{
	DokumenKindSuratPernyataan: true,
	DokumenKindKTP:             true,
	DokumenKindKK:              true,
	DokumenKindMutasi:          true,
	DokumenKindPembayaran:      true,
}

type DokumenStatus string

const (
	DokumenStatusPending  DokumenStatus = "pending"
	DokumenStatusVerified DokumenStatus = "verified"
	DokumenStatusRejected DokumenStatus = "rejected"
)

type SantriRequestStatus string

const (
	SantriRequestStatusPending  SantriRequestStatus = "pending"
	SantriRequestStatusApproved SantriRequestStatus = "approved"
	SantriRequestStatusRejected SantriRequestStatus = "rejected"
)

const (
	CodeInvalidNISFormat domainerr.Code = "DOMAIN_INVALID_NIS_FORMAT"

	CodeSantriNotFound          domainerr.Code = "DOMAIN_SANTRI_NOT_FOUND"
	CodeSantriPersistenceFailed domainerr.Code = "DOMAIN_SANTRI_PERSISTENCE_FAILED"
	CodeSantriQueryFailed       domainerr.Code = "DOMAIN_SANTRI_QUERY_FAILED"
	CodeSantriDuplicate         domainerr.Code = "DOMAIN_SANTRI_DUPLICATE"

	CodeDokumenNotFound          domainerr.Code = "DOMAIN_SANTRI_DOKUMEN_NOT_FOUND"
	CodeDokumenPersistenceFailed domainerr.Code = "DOMAIN_SANTRI_DOKUMEN_PERSISTENCE_FAILED"
	CodeDokumenQueryFailed       domainerr.Code = "DOMAIN_SANTRI_DOKUMEN_QUERY_FAILED"
	CodeDokumenInvalidKind       domainerr.Code = "DOMAIN_SANTRI_DOKUMEN_INVALID_KIND"
	CodeDokumenInvalidStatus     domainerr.Code = "DOMAIN_SANTRI_DOKUMEN_INVALID_STATUS"

	CodeSantriRequestNotFound          domainerr.Code = "DOMAIN_SANTRI_REQUEST_NOT_FOUND"
	CodeSantriRequestPersistenceFailed domainerr.Code = "DOMAIN_SANTRI_REQUEST_PERSISTENCE_FAILED"
	CodeSantriRequestQueryFailed       domainerr.Code = "DOMAIN_SANTRI_REQUEST_QUERY_FAILED"
	CodeSantriRequestAlreadyExists     domainerr.Code = "DOMAIN_SANTRI_REQUEST_ALREADY_EXISTS"
	CodeSantriRequestInvalidStatus     domainerr.Code = "DOMAIN_SANTRI_REQUEST_INVALID_STATUS"
)

var AllowedContentTypes = map[string]bool{
	"image/jpeg":          true,
	"image/png":           true,
	"application/pdf":     true,
}
