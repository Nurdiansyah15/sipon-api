package minio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"sipon-api/internal/app/port"
)

type MinioFileUploader struct {
	client        *minio.Client
	bucket        string
	privateBucket string
	endpoint      string
	useSSL        bool
}

func NewMinioFileUploader(endpoint, accessKey, secretKey, bucket, privateBucket string, useSSL bool) (*MinioFileUploader, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: gagal membuat client: %w", err)
	}

	return &MinioFileUploader{
		client:        client,
		bucket:        bucket,
		privateBucket: privateBucket,
		endpoint:      endpoint,
		useSSL:        useSSL,
	}, nil
}

func (u *MinioFileUploader) resolveBucket(privacy port.PrivacyRule) string {
	if privacy == port.PrivacyPrivate {
		return u.privateBucket
	}
	return u.bucket
}

func (u *MinioFileUploader) RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy port.PrivacyRule) (string, string, string, error) {
	bucket := u.resolveBucket(privacy)
	key := strings.TrimPrefix(objectName, "/")

	presignURL, err := u.client.PresignedPutObject(ctx, bucket, key, expiry)
	if err != nil {
		return "", "", "", fmt.Errorf("minio: gagal membuat presigned URL: %w", err)
	}

	var publicURL string
	if privacy == port.PrivacyPublic {
		publicURL = u.PublicURL(objectName)
	}

	return presignURL.String(), objectName, publicURL, nil
}

func (u *MinioFileUploader) ConfirmUpload(ctx context.Context, key string) error {
	// Untuk flow dengan presigned URL, tidak perlu aksi konfirmasi khusus
	// karena bucket public sudah anonymous-readable.
	// Method ini ada untuk memenuhi interface dan untuk flow di mana
	// status upload perlu di-track secara eksplisit.
	return nil
}

func (u *MinioFileUploader) MarkDeleted(ctx context.Context, key string) error {
	cleaned := strings.TrimPrefix(key, "/")
	err := u.client.RemoveObject(ctx, u.bucket, cleaned, minio.RemoveObjectOptions{ForceDelete: true})
	if err != nil {
		// Best-effort — file mungkin belum ada atau sudah dihapus
		// Cek apakah error bukan "not found" baru return error
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil
		}
		return fmt.Errorf("minio: gagal menghapus objek: %w", err)
	}
	return nil
}

func (u *MinioFileUploader) PublicURL(key string) string {
	cleaned := strings.TrimPrefix(key, "/")
	scheme := "http"
	if u.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, u.endpoint, u.bucket, cleaned)
}

func (u *MinioFileUploader) KeyFromURL(url string) string {
	prefix := fmt.Sprintf("%s://%s/%s/", "http", u.endpoint, u.bucket)
	scheme := "http"
	if u.useSSL {
		scheme = "https"
		prefix = fmt.Sprintf("%s://%s/%s/", scheme, u.endpoint, u.bucket)
	}
	if strings.HasPrefix(url, prefix) {
		return "/" + strings.TrimPrefix(url, prefix)
	}
	// Coba parse generic — ambil path setelah bucket terakhir
	if idx := strings.LastIndex(url, u.bucket); idx != -1 {
		afterBucket := url[idx+len(u.bucket):]
		return "/" + strings.TrimPrefix(afterBucket, "/")
	}
	return url
}

func (u *MinioFileUploader) DeleteObject(ctx context.Context, key string, privacy port.PrivacyRule) error {
	bucket := u.resolveBucket(privacy)
	cleaned := strings.TrimPrefix(key, "/")
	err := u.client.RemoveObject(ctx, bucket, cleaned, minio.RemoveObjectOptions{ForceDelete: true})
	if err != nil {
		return fmt.Errorf("minio: gagal menghapus objek: %w", err)
	}
	return nil
}
