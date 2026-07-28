package authUsecase

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	media "sipon-api/internal/app/service/media"
)

// Content type yang diizinkan untuk avatar (CLAUDE.md §9: gambar).
var avatarAllowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// extByContentType memetakan content-type ke ekstensi file.
var extByContentType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type AvatarPresignUseCase struct {
	fileUploader port.FileUploader
}

func NewAvatarPresignUseCase(fileUploader port.FileUploader) *AvatarPresignUseCase {
	return &AvatarPresignUseCase{fileUploader: fileUploader}
}

// Required — role: any | perm: - | benefit: -
func (uc *AvatarPresignUseCase) Execute(ctx context.Context, req dto.AvatarPresignRequest) (*dto.AvatarPresignResponse, error) {
	ct := strings.TrimSpace(strings.ToLower(req.ContentType))
	if !avatarAllowedContentTypes[ct] {
		return nil, apperror.Unprocessable(string(apperror.CodeContentTypeNotAllowed), nil)
	}

	ext := extByContentType[ct]
	objectName := path.Join(string(media.ObjectPathAvatar), uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, media.AvatarPresignExpiry, port.PrivacyPublic)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), fmt.Errorf("gagal membuat presigned URL: %w", err))
	}

	return &dto.AvatarPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(media.AvatarPresignExpiry.Seconds()),
	}, nil
}
