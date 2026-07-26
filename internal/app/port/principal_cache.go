package port

import (
	"context"
	"sipon-api/internal/app/service/principal"
	"time"
)

type PrincipalCachePort interface {
	Get(ctx context.Context, userID string) (*principal.Principal, error)
	Set(ctx context.Context, userID string, p *principal.Principal, ttl time.Duration) error
	Delete(ctx context.Context, userID string) error
}
