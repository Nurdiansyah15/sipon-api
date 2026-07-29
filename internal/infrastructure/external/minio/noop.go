package minio

import (
	"context"
	"time"

	"sipon-api/internal/app/port"
)

type NoopFileUploader struct{}

func NewNoopFileUploader() *NoopFileUploader {
	return &NoopFileUploader{}
}

func (n *NoopFileUploader) RequestUpload(_ context.Context, objectName, _ string, _ time.Duration, _ port.PrivacyRule) (string, string, string, error) {
	return "", objectName, "", nil
}

func (n *NoopFileUploader) ConfirmUpload(_ context.Context, _ string) error {
	return nil
}

func (n *NoopFileUploader) MarkDeleted(_ context.Context, _ string) error {
	return nil
}

func (n *NoopFileUploader) PublicURL(key string) string {
	return key
}

func (n *NoopFileUploader) KeyFromURL(url string) string {
	return url
}

func (n *NoopFileUploader) GeneratePresignedDownloadURL(_ context.Context, key string, _ port.PrivacyRule, _ time.Duration) (string, error) {
	return key, nil
}

func (n *NoopFileUploader) DeleteObject(_ context.Context, _ string, _ port.PrivacyRule) error {
	return nil
}
