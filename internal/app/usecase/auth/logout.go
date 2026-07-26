package authUsecase

import (
	"context"
	"time"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
)

type LogoutUseCase struct {
	revocationStore port.SessionRevocationStore
	accessTokenTTL  time.Duration
}

func NewLogoutUseCase(revocationStore port.SessionRevocationStore, accessTokenTTL time.Duration) *LogoutUseCase {
	return &LogoutUseCase{revocationStore: revocationStore, accessTokenTTL: accessTokenTTL}
}

// Required — role: any | perm: - | benefit: -
// Merevoke session (access token) sekarang. TTL revoke = access token TTL —
// setelah itu token akan expired sendiri jadi tidak perlu diingat lebih lama.
func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	if err := uc.revocationStore.RevokeSession(ctx, sessionID, uc.accessTokenTTL); err != nil {
		return apperror.Internal(string(apperror.CodeInternal), err)
	}
	return nil
}
