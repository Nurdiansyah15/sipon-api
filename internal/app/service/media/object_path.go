package media

import "time"

// ObjectPath adalah tipe yang merepresentasikan prefix path objek di object storage.
// Domain entity menyimpan key, bukan full URL.
type ObjectPath string

const (
	ObjectPathAvatar ObjectPath = "/avatars/"
)

// AvatarPresignExpiry — durasi presigned URL untuk upload avatar (5 menit).
const AvatarPresignExpiry = 5 * time.Minute
