package testhelper

import (
	"context"
	"sync"
	"time"

	"sipon-api/internal/app/service/principal"
)

// ── noopPrincipalCache ────────────────────────────────────────────────────────
// Selalu cache miss — principal dimuat dari DB setiap request.
type noopPrincipalCache struct{}

func (noopPrincipalCache) Get(_ context.Context, _ string) (*principal.Principal, error) {
	return nil, nil
}
func (noopPrincipalCache) Set(_ context.Context, _ string, _ *principal.Principal, _ time.Duration) error {
	return nil
}
func (noopPrincipalCache) Delete(_ context.Context, _ string) error { return nil }

// ── noopSMSSender ─────────────────────────────────────────────────────────────
type noopSMSSender struct{}

func (noopSMSSender) SendOTP(_, _ string) error { return nil }

// ── fakeSessionRevocationStore ────────────────────────────────────────────────
// Implementasi in-memory (bukan Redis) untuk port.SessionRevocationStore.
// Harus benar-benar menyimpan state, karena test butuh assert token lama
// ditolak setelah logout/logout-all.
type fakeSessionRevocationStore struct {
	mu                  sync.Mutex
	revokedSess         map[string]struct{}
	revokedBefore       map[string]time.Time
	deviceRevokedBefore map[string]time.Time
}

func newFakeSessionRevocationStore() *fakeSessionRevocationStore {
	return &fakeSessionRevocationStore{
		revokedSess:         map[string]struct{}{},
		revokedBefore:       map[string]time.Time{},
		deviceRevokedBefore: map[string]time.Time{},
	}
}

func deviceRevocationKey(userID, deviceID string) string {
	return userID + ":" + deviceID
}

func (s *fakeSessionRevocationStore) RevokeSession(_ context.Context, sessionID string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revokedSess[sessionID] = struct{}{}
	return nil
}

func (s *fakeSessionRevocationStore) IsSessionRevoked(_ context.Context, sessionID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.revokedSess[sessionID]
	return ok, nil
}

func (s *fakeSessionRevocationStore) RevokeAllBefore(_ context.Context, userID string, before time.Time, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revokedBefore[userID] = before
	return nil
}

func (s *fakeSessionRevocationStore) RevokedBefore(_ context.Context, userID string) (*time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.revokedBefore[userID]
	if !ok {
		return nil, nil
	}
	return &t, nil
}

func (s *fakeSessionRevocationStore) RevokeDeviceBefore(_ context.Context, userID, deviceID string, before time.Time, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceRevokedBefore[deviceRevocationKey(userID, deviceID)] = before
	return nil
}

func (s *fakeSessionRevocationStore) DeviceRevokedBefore(_ context.Context, userID, deviceID string) (*time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.deviceRevokedBefore[deviceRevocationKey(userID, deviceID)]
	if !ok {
		return nil, nil
	}
	return &t, nil
}
