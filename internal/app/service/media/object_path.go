package media

import "time"

// ObjectPath adalah tipe yang merepresentasikan prefix path objek di object storage.
// Domain entity menyimpan key, bukan full URL.
type ObjectPath string

const (
	ObjectPathAvatar        ObjectPath = "/avatars/"
	ObjectPathSantriDokumen ObjectPath = "/santri/dokumen/"
)

// AvatarPresignExpiry — durasi presigned URL untuk upload avatar (5 menit).
const AvatarPresignExpiry = 5 * time.Minute

// SantriDokumenPresignUploadExpiry — durasi presigned URL untuk upload dokumen santri (10 menit).
const SantriDokumenPresignUploadExpiry = 10 * time.Minute

// SantriDokumenAccessTTL — durasi presigned GET URL untuk akses dokumen santri (15 menit).
const SantriDokumenAccessTTL = 15 * time.Minute
