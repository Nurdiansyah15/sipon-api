package port

import (
	"context"
	"time"
)

type PrivacyRule string

const (
	PrivacyPublic  PrivacyRule = "PUBLIC"
	PrivacyPrivate PrivacyRule = "PRIVATE"
)

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	MarkDeleted(ctx context.Context, key string) error
	PublicURL(key string) string
	KeyFromURL(url string) string
}
