package port

import (
	"context"
	"time"
)

// SessionRevocationStore memberi kapabilitas revocation nyata untuk JWT yang
// secara desain stateless (lihat TokenGenerator) — dipakai untuk enforce
// logout (per-session, via RevokeSession) dan logout-all (revoke semua token
// yang diterbitkan sebelum suatu waktu, via RevokeAllBefore), tanpa perlu
// tabel session/refresh-token baru. Cukup memperluas Redis yang sudah dipakai
// untuk principal cache (lihat PrincipalCachePort).
type SessionRevocationStore interface {
	// RevokeSession menandai satu access token session sebagai revoked selama
	// ttl (biasanya = access token TTL — setelah itu token expired sendiri).
	RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error
	// IsSessionRevoked mengecek apakah sessionID sudah di-revoke.
	IsSessionRevoked(ctx context.Context, sessionID string) (bool, error)
	// RevokeAllBefore menandai semua token (access & refresh) milik userID yang
	// diterbitkan sebelum `before` sebagai tidak valid lagi, selama ttl
	// (biasanya = refresh token TTL, karena itu yang berumur paling panjang).
	RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error
	// RevokedBefore mengembalikan timestamp revoke-all terakhir untuk userID,
	// nil kalau belum pernah logout-all atau sudah expired dari store.
	RevokedBefore(ctx context.Context, userID string) (*time.Time, error)

	// RevokeDeviceBefore menandai semua token (access & refresh) yang membawa
	// deviceID ini milik userID, diterbitkan sebelum `before`, sebagai tidak
	// valid lagi selama ttl (biasanya = refresh token TTL). Dipakai untuk
	// "logout device lain" (bukan device yang sedang request) — device itu
	// sendiri yang tahu sessionID-nya saat request, tapi device LAIN tidak
	// bisa direvoke lewat RevokeSession karena sessionID di-mint ulang tiap
	// refresh dan tidak pernah disimpan terkait ke device manapun. deviceID
	// (device_registrations.device_id) adalah identitas yang stabil, jadi
	// dipakai sebagai kunci revocation di sini — hanya berlaku kalau device
	// itu mengirim device_id saat login (lihat dto.LoginRequest.DeviceID).
	RevokeDeviceBefore(ctx context.Context, userID, deviceID string, before time.Time, ttl time.Duration) error
	// DeviceRevokedBefore mengembalikan timestamp revoke-device terakhir untuk
	// userID+deviceID, nil kalau belum pernah di-revoke atau sudah expired.
	DeviceRevokedBefore(ctx context.Context, userID, deviceID string) (*time.Time, error)
}
