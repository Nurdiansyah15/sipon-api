package rolepermission

import (
	"context"

	"sipon-api/internal/app/dto"
	roleconstant "sipon-api/internal/domain/role/constant"
)

// ListPermissionKeysUseCase mengembalikan katalog permission yang dikenal
// sistem (dari constant.AllPermissionDefinitions) — dipakai frontend untuk
// menampilkan pilihan permission valid saat assign ke custom role, tanpa
// hardcode di client.
type ListPermissionKeysUseCase struct{}

func NewListPermissionKeysUseCase() *ListPermissionKeysUseCase {
	return &ListPermissionKeysUseCase{}
}

// Required — role: superadmin, usergod | perm: - | benefit: -
func (uc *ListPermissionKeysUseCase) Execute(_ context.Context) []dto.PermissionKeyResponse {
	defs := roleconstant.AllPermissionDefinitions()
	items := make([]dto.PermissionKeyResponse, 0, len(defs))
	for _, d := range defs {
		items = append(items, dto.PermissionKeyResponse{
			Key:         string(d.Key),
			DisplayName: d.DisplayName,
			Description: d.Description,
		})
	}
	return items
}
