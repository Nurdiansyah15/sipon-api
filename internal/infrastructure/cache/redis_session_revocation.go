package cache

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSessionRevocationStore mengimplementasikan port.SessionRevocationStore
// via tiga pola key Redis:
//   - "session:revoked:<sessionID>"           -> revoke satu session (logout device sekarang)
//   - "user:revoked_before:<userID>"           -> unix timestamp revoke-all terakhir (logout semua device)
//   - "device:revoked_before:<userID>:<devID>" -> unix timestamp revoke-device terakhir (logout device lain)
type RedisSessionRevocationStore struct {
	client *redis.Client
}

func NewRedisSessionRevocationStore(client *redis.Client) *RedisSessionRevocationStore {
	return &RedisSessionRevocationStore{client: client}
}

func (s *RedisSessionRevocationStore) sessionKey(sessionID string) string {
	return "session:revoked:" + sessionID
}

func (s *RedisSessionRevocationStore) userRevokedBeforeKey(userID string) string {
	return "user:revoked_before:" + userID
}

func (s *RedisSessionRevocationStore) deviceRevokedBeforeKey(userID, deviceID string) string {
	return "device:revoked_before:" + userID + ":" + deviceID
}

func (s *RedisSessionRevocationStore) RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	if sessionID == "" {
		return nil
	}
	return s.client.Set(ctx, s.sessionKey(sessionID), "1", ttl).Err()
}

func (s *RedisSessionRevocationStore) IsSessionRevoked(ctx context.Context, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}
	n, err := s.client.Exists(ctx, s.sessionKey(sessionID)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *RedisSessionRevocationStore) RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error {
	if userID == "" {
		return nil
	}
	return s.client.Set(ctx, s.userRevokedBeforeKey(userID), strconv.FormatInt(before.Unix(), 10), ttl).Err()
}

func (s *RedisSessionRevocationStore) RevokedBefore(ctx context.Context, userID string) (*time.Time, error) {
	if userID == "" {
		return nil, nil
	}
	raw, err := s.client.Get(ctx, s.userRevokedBeforeKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(sec, 0)
	return &t, nil
}

func (s *RedisSessionRevocationStore) RevokeDeviceBefore(ctx context.Context, userID, deviceID string, before time.Time, ttl time.Duration) error {
	if userID == "" || deviceID == "" {
		return nil
	}
	return s.client.Set(ctx, s.deviceRevokedBeforeKey(userID, deviceID), strconv.FormatInt(before.Unix(), 10), ttl).Err()
}

func (s *RedisSessionRevocationStore) DeviceRevokedBefore(ctx context.Context, userID, deviceID string) (*time.Time, error) {
	if userID == "" || deviceID == "" {
		return nil, nil
	}
	raw, err := s.client.Get(ctx, s.deviceRevokedBeforeKey(userID, deviceID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(sec, 0)
	return &t, nil
}
