# CLAUDE.md

Panduan ini dibaca otomatis oleh Claude Code setiap sesi dan mencakup arsitektur, konvensi, dan aturan pengembangan untuk backend berbasis DDD (Domain-Driven Design).

---

## 0. Teori Umum DDD (Ringkas)

DDD (Domain-Driven Design) adalah pendekatan desain software yang menempatkan domain bisnis sebagai pusat keputusan desain.

Konsep umum yang dipakai:

- **Ubiquitous Language**: istilah domain di diskusi, dokumen, dan kode harus konsisten.
- **Bounded Context**: batas model domain agar istilah dan aturan tidak bercampur antar area.
- **Entity**: objek domain yang identitasnya penting dan punya siklus hidup.
- **Value Object**: objek tanpa identitas, dibandingkan berdasarkan nilai.
- **Aggregate**: kumpulan object domain dengan satu aggregate root sebagai pintu perubahan.
- **Repository**: abstraksi akses aggregate dari/ke storage.
- **Domain Service**: operasi domain yang tidak pas berada pada satu entity.
- **Application/Usecase Service**: orkestrasi flow bisnis lintas domain object dan port.

Prinsip ringkas:

- Domain model berisi rule bisnis inti.
- Usecase mengatur alur, bukan menyimpan rule inti.
- Infrastruktur hanya detail teknis implementasi.

---

## 1. Alur Standar Pembuatan Endpoint

Gunakan alur berikut secara konsisten:

```
HTTP Handler -> Usecase -> Domain (Repository Interface / Entity / Domain Service) -> Model Persistence
```

Catatan implementasi:

- Handler ada di layer `interfaces/http`.
- Usecase ada di layer `internal/app/usecase`.
- Domain ada di layer `internal/domain`.
- Implementasi repository ada di `internal/infrastructure/persistence`.

---

## 2. Tanggung Jawab Tiap Layer

### Handler

- Terima request HTTP.
- Parsing dan validasi dasar payload/param.
- Panggil satu usecase yang sesuai.
- Mapping hasil usecase ke response HTTP.
- Tidak berisi logika bisnis domain.

### Usecase

- Menjadi orkestrator proses bisnis.
- Mengatur urutan eksekusi aturan domain.
- Orkestrasi diutamakan melalui domain method, bukan logika bisnis manual di usecase.
- Bisa mengambil domain object dari repository.
- Bisa membuat entity baru melalui constructor/factory domain.
- Bisa memanggil domain service jika rule lintas entity/aggregate.
- Tidak berisi detail SQL, query builder, atau akses driver DB.
- Penulisan usecase wajib satu file untuk satu usecase.
- Wajib memetakan domain object ke DTO sebelum mengembalikan hasil ke handler. Handler tidak boleh menerima domain object mentah.

### Domain

- Menyimpan aturan bisnis inti: entity, value object, domain service.
- Mendefinisikan kontrak repository (interface), bukan implementasi.
- Tidak boleh import package `infrastructure` atau `interfaces`.

### Domain Service

- Digunakan untuk menampung rule bisnis yang tidak cocok ditempatkan pada satu entity atau value object.
- Biasanya menangani proses yang melibatkan beberapa entity, aggregate, atau operasi domain yang membutuhkan koordinasi lebih dari satu objek domain.
- Tetap merupakan bagian dari layer domain dan berisi logika bisnis, bukan detail teknis.
- Boleh bergantung pada repository interface yang berada di domain untuk mengambil atau memvalidasi data yang dibutuhkan oleh rule bisnis.
- Tidak boleh bergantung pada implementasi repository, database driver, HTTP framework, atau komponen infrastructure lainnya.
- Dapat dipanggil oleh usecase maupun oleh domain object lain jika diperlukan.

**Contoh penggunaan:**

- Validasi apakah sebuah aktor dapat bergabung ke suatu scope berdasarkan aturan domain tertentu.
- Perhitungan hak akses efektif yang berasal dari kombinasi beberapa role.
- Penentuan role/state default ketika anggota baru masuk ke suatu scope.
- Rule yang membutuhkan data dari beberapa aggregate berbeda sebelum menghasilkan keputusan domain.

**Prinsip utama:**

Jika sebuah aturan bisnis tidak memiliki "rumah yang jelas" pada satu entity karena membutuhkan koordinasi beberapa objek domain atau akses repository domain, maka aturan tersebut merupakan kandidat untuk ditempatkan pada domain service.

Alur yang umum:

```text
Usecase
   ↓
Domain Service
   ↓
Repository Interface (Domain)
   ↓
Entity / Value Object
```

Domain service tetap berada di pusat logika bisnis. Repository hanya digunakan sebagai sumber data yang diperlukan untuk menjalankan rule domain tersebut.

### Infrastructure Persistence

- Mengimplementasikan interface repository dari domain.
- Menangani detail teknis penyimpanan data (SQL, transaction, mapping model).
- Mengubah data storage menjadi domain object dan sebaliknya.

### Model Persistence

- Struktur data yang dipakai untuk simpan/ambil ke DB.
- Hanya untuk kebutuhan persistence, bukan sumber aturan bisnis.

---

## 3. Dependency Rule (Wajib)

Dependensi hanya boleh mengarah ke dalam:

```
interfaces -> app/usecase -> domain
infrastructure -> domain
```

Larangan:

- Domain import infrastructure.
- Usecase import implementasi repository langsung.
- Handler akses DB langsung tanpa usecase.

---

## 4. Pola Kerja Usecase terhadap Domain

Di dalam usecase, pola umum:

1. Ambil data domain dari repository bila dibutuhkan.
2. Buat entity baru bila proses membutuhkan objek baru.
3. Jalankan domain method (entity/value object) sebagai jalur utama eksekusi rule bisnis.
4. Pakai domain service jika aturan tidak cocok ditempatkan pada satu entity.
5. Simpan perubahan melalui repository interface.
6. Petakan domain object ke DTO sebagai langkah terakhir sebelum return. Usecase tidak boleh mengembalikan domain object langsung ke caller.

Hindari memindahkan rule domain ke usecase dalam bentuk if/else panjang. Jika rule berulang atau kompleks, pindahkan ke domain method atau domain service.

Usecase adalah tempat orkestrasi, domain adalah sumber kebenaran rule bisnis.

---

## 5. Lokasi Kontrak dan Implementasi Repository

- Kontrak/interface repository: `internal/domain/...` (atau `internal/app/port` bila port aplikasi).
- Implementasi repository: `internal/infrastructure/persistence/...`.

Prinsip utama:

- Kode domain dan usecase bergantung pada interface.
- Detail persistence dapat diganti tanpa mengubah rule domain/usecase.

### 5.1 File Layout untuk Implementasi Persistence

- **Pisahkan** implementasi repository dan query model ke file terpisah. Untuk setiap bounded context/paket persistence, tempatkan implementasi repositori (CRUD untuk aggregate) di satu atau beberapa file bertipe `postgres_<context>_repository.go` dan implementasi query model (listing/pagination/analytics kompleks) di file `postgres_<context>_query.go`.
- **Contoh penamaan**: `postgres_<context>_repository.go`, `postgres_<context>_query.go` untuk setiap bounded context (mis. sebuah context bernama `role_permission` akan punya `postgres_role_permission_repository.go` dan `postgres_role_permission_query.go`).
- **Alasan**: memisahkan concern membuat maintenance lebih mudah, memudahkan review untuk query kompleks, dan konsisten dengan pola `port`/`QueryReadModel` yang dipakai pada layer aplikasi.
- **Aturan singkat**:
  - Implementasi method yang memetakan entity domain langsung (Save/Update/FindByID/Delete/List sederhana) adalah tanggung jawab _repository implementation_ dan harus berada di file `postgres_<context>_repository.go`.
  - Implementasi method yang mengembalikan `*port.XxxListReadResult`, melakukan join kompleks, aggregation, pagination, atau analytics adalah _query model_ dan harus berada di file `postgres_<context>_query.go`.
  - Nama tipe struct tetap boleh mencerminkan perannya, mis. `Postgres<Context>Repository` untuk repository dan `Postgres<Context>Query` untuk query model.

---

## 6. Aturan Pemakaian Port Query Model

Port query model dipakai hanya untuk kebutuhan query yang memang butuh pagination atau listing/filtering/sorting yang kompleks.

Gunakan domain repository langsung untuk query sederhana, contohnya:

- get by id
- get by unique field
- query sederhana lain yang masih representasi aggregate/domain object

Jangan membuat query model hanya untuk operasi baca sederhana. Jika belum ada kebutuhan pagination/listing kompleks, tetap gunakan domain repository.

---

## 7. Aturan Error Mapping di Persistence (Repository Implementation)

Implementasi repository untuk bounded context apapun WAJIB memetakan error storage ke domain error code yang sesuai. Jangan expose `sql.ErrNoRows` atau `pgconn.PgError` mentah ke usecase.

**Aturan:**

- `sql.ErrNoRows` → `domainerr.New(constant.CodeXxxNotFound, "...")` — selalu gunakan not-found code yang tepat per entitas.
- Error INSERT/UPDATE/DELETE → `domainerr.Wrap(constant.CodeXxxPersistenceFailed, "...", err)` — untuk kegagalan penyimpanan umum.
- Error SELECT/listing → `domainerr.Wrap(constant.CodeXxxQueryFailed, "...", err)` — untuk kegagalan query/listing.
- Unique constraint violation (pgconn.PgError Code `"23505"`) → `domainerr.New(constant.CodeXxxDuplicateKey, "...")` — cek menggunakan `errors.As(err, &pgErr)`.

**Contoh:**

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    return domainerr.New(constant.CodeXxxDuplicateKey, "...")
}
return domainerr.Wrap(constant.CodeXxxPersistenceFailed, msg, err)
```

Usecase memetakan domain error ke AppError melalui helper (misalnya `mapXxxDomainError()`). Persistence hanya boleh mengembalikan domain error, bukan AppError.

---

## 8. Pola Return untuk Endpoint dengan Pagination Meta

Untuk endpoint listing dengan pagination, gunakan tipe-tipe berikut yang sudah ada di package `dto`:

```go
// dto/pagination_dto.go — tipe standar untuk semua bounded context
type PaginationParams struct {
    Page     *int
    Limit    *int
    SortBy   *string
    SortType *string
}

type Meta struct {
    CurrentPage int64 `json:"current_page"`
    PerPage     int64 `json:"per_page"`
    Total       int64 `json:"total"`
    TotalPages  int64 `json:"total_pages"`
}
```

**Port ReadQuery** — embed `PaginationParams`:

```go
type XxxListReadQuery struct {
    // filter fields...
    PaginationParams
}
```

**Port ReadResult** — sertakan `Meta Meta`:

```go
type XxxListReadResult struct {
    Items []XxxReadItem
    Meta  Meta
}
```

**DTO Query struct** — field pagination sebagai pointer (agar bisa nil/opsional):

```go
type ListXxxQuery struct {
    // filter fields...
    Page     *int    `form:"page"`
    Limit    *int    `form:"limit"`
    SortBy   *string `form:"sort_by"`
    SortType *string `form:"sort_type"`
}
```

**Usecase Execute()** — wajib mengembalikan data, meta, dan error sebagai **tiga return value terpisah**:

```go
// BENAR
func (uc *XxxUseCase) Execute(ctx context.Context, req dto.ListXxxQuery) ([]dto.XxxItem, port.Meta, error)

// SALAH — jangan bungkus data+meta dalam satu DTO response struct
func (uc *XxxUseCase) Execute(ctx context.Context, req dto.ListXxxQuery) (*dto.XxxListResponse, error)
```

Mapping DTO → port PaginationParams (langsung karena field sudah pointer):

```go
PaginationParams: port.PaginationParams{
    Page:     req.Page,
    Limit:    req.Limit,
    SortBy:   req.SortBy,
    SortType: req.SortType,
}
```

**Persistence** — gunakan helper `resolvePaginationParams()` untuk normalisasi dan kalkulasi:

```go
limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(
    query.PaginationParams, defaultLimit, maxLimit, sortableMap, defaultSort, defaultSortType,
)
totalPages := (total + int64(limit) - 1) / int64(limit)
return &port.XxxListReadResult{
    Items: items,
    Meta:  port.Meta{CurrentPage: currentPage, PerPage: int64(limit), Total: total, TotalPages: totalPages},
}, nil
```

**Handler** — gunakan `respond.SuccessWithMeta`, bukan `respond.OK`, untuk endpoint listing:

```go
data, meta, err := h.useCases.XxxList.Execute(c.Request.Context(), req)
if err != nil { ... }
respond.SuccessWithMeta(c, 200, "...", data, meta)
```

---

## 9. Aturan Media Upload (Presign / Confirm / Delete)

Setiap titik upload media mengikuti pola berikut secara konsisten di seluruh modul.

### Alur Upload

```
Client                 API                       Object Storage
  |                     |                           |
  |-- POST /presign --> |                           |
  |                     |-- RequestUpload -------> |
  |<-- presign_url, key |                           |
  |                     |                           |
  |-- PUT presign_url --|------------------------> |  (langsung ke storage)
  |                     |                           |
  |-- (opsional) POST /confirm?key=... ----------> |  ← konfirmasi eksplisit
  |   atau kirim key di payload bisnis             |  ← konfirmasi implisit
```

### Dua Cara Konfirmasi

**1. Eksplisit via `/confirm?key=...`**
Digunakan ketika media di-upload sebelum entitas bisnis dibuat/diperbarui. Client melakukan konfirmasi terpisah agar file tidak otomatis dihapus sebelum digunakan.

**2. Implisit via payload bisnis (create/update)**
Jika `image_url`, `video_url`, atau field media lain diisi dalam payload create/update, usecase akan memanggil `confirmXxxMediaKey()` (atau helper serupa) setelah entitas berhasil disimpan. Tidak perlu panggil `/confirm` tersendiri.

Kedua cara tersebut bisa dipakai — client bisa memilih sesuai flow UI-nya.

### Implementasi di Usecase

```go
// Setelah repo.Save / repo.Update sukses:
confirmXxxMediaKey(ctx, uc.mediaUploadSvc, entity.ImageURL)
confirmXxxMediaKey(ctx, uc.mediaUploadSvc, entity.VideoURL)
// dst untuk setiap field media

// Untuk update — hapus media lama jika diganti:
markXxxMediaDeleted(ctx, uc.mediaUploadSvc, oldImageURL, entity.ImageURL)
```

Helper `confirmXxxMediaKey` dan `markXxxMediaDeleted` ada di `helpers.go` masing-masing paket usecase. Keduanya best-effort (error diabaikan) agar tidak menggagalkan operasi bisnis.

### Port yang Digunakan

```go
type MediaUploadService interface {
    RequestUpload(ctx, objectName, contentType, expiry, privacy) (presignURL, key, publicURL, error)
    ConfirmUpload(ctx, key) error
    MarkDeleted(ctx, key) error
    NormalizeKey(rawURL) string
    PublicURL(key) string
}
```

`NormalizeKey` mengekstrak object key dari URL (menghapus domain storage). Selalu panggil ini sebelum meneruskan ke `ConfirmUpload` / `MarkDeleted`.

### ObjectPath Constants

Semua path objek didefinisikan di `internal/app/service/media/object_path.go`, satu constant per jenis media per modul, contoh:

```go
ObjectPathXxxImage ObjectPath = "/xxx/images/"
ObjectPathXxxVideo ObjectPath = "/xxx/videos/"
```

### Endpoint Pattern

Setiap modul dengan media memiliki 6 endpoint bila punya gambar dan video:

| Method | Path | Keterangan |
|--------|------|------------|
| POST | `/<module>/media/image/presign` | Minta presigned URL upload gambar |
| POST | `/<module>/media/image/confirm` | Konfirmasi gambar sudah diupload |
| DELETE | `/<module>/media/image` | Tandai gambar sebagai deleted |
| POST | `/<module>/media/video/presign` | Minta presigned URL upload video |
| POST | `/<module>/media/video/confirm` | Konfirmasi video sudah diupload |
| DELETE | `/<module>/media/video` | Tandai video sebagai deleted |

Untuk modul lain yang hanya punya satu jenis media (gambar), cukup 3 endpoint: `presign`, `confirm`, `delete`.

`confirm` dan `delete` menerima `key` via query param: `?key=<module>/images/uuid.jpg`.

### Presign Request / Response DTO

```go
type XxxMediaPresignRequest struct {
    ContentType string `json:"content_type" binding:"required"`
}
type XxxMediaPresignResponse struct {
    PresignURL string `json:"presign_url"`
    Key        string `json:"key"`
    PublicURL  string `json:"public_url"`
    ExpiresIn  int    `json:"expires_in"`
}
type XxxMediaConfirmResponse struct {
    PublicURL string `json:"public_url"`
}
```

### Content Type yang Diizinkan

- **Gambar**: `image/jpeg`, `image/png`, `image/webp`, `image/gif`
- **Video**: `video/mp4`, `video/webm`, `video/quicktime`

Validasi dilakukan di usecase presign, bukan di handler.

---

## 10. Aturan Build URL untuk Field Media di DTO

### Prinsip Utama

**Domain entity menyimpan key (object path), bukan full URL.**

Contoh: `entity.AvatarURL.Value()` → `"/avatars/uuid.jpg"` (key), bukan `"https://cdn.example.com/avatars/uuid.jpg"`.

**Setiap field URL di DTO wajib dikonversi dari key ke public URL di dalam usecase**, menggunakan `fileUploader.PublicURL(key)` sebelum dikembalikan ke handler. Handler tidak boleh menerima key mentah.

### Port yang Digunakan

```go
// port.FileUploader — injected ke usecase via constructor
type FileUploader interface {
    PublicURL(key string) string       // key → full public URL
    KeyFromURL(url string) string      // full URL → key (kebalikannya)
    // ... metode lain untuk presign/upload/delete
}
```

`port.MediaUploadService` juga memiliki `PublicURL(key string) string` yang mendelegasikan ke `FileUploader`. Keduanya boleh dipakai tergantung dependency yang tersedia di usecase.

### Pola Implementasi

**Inline (untuk field tunggal sederhana):**

```go
v := strings.TrimSpace(entity.AvatarURL.Value())
if uc.fileUploader != nil {
    v = uc.fileUploader.PublicURL(v)
}
avatar = &v
```

**Via helper (untuk field nullable/pointer, dipakai berulang):**

```go
// Pola helper yang konsisten di semua paket usecase
func resolveXxxImageURL(fileUploader port.FileUploader, key *string) *string {
    if key == nil {
        return nil
    }
    v := strings.TrimSpace(*key)
    if v == "" {
        return nil
    }
    if strings.Contains(v, "://") {
        return &v  // sudah full URL, lewati konversi
    }
    url := fileUploader.PublicURL(v)
    return &url
}
```

Helper ini ada di `helpers.go` masing-masing paket usecase yang memiliki field media.

### Aturan

1. **Usecase yang mengembalikan DTO dengan field URL** (avatar, image_url, video_url, cover_url, dst.) **wajib inject `fileUploader port.FileUploader`** (atau `mediaUploadSvc port.MediaUploadService`) sebagai dependency.

2. **Setiap field URL di DTO wajib dikonversi** via `fileUploader.PublicURL(key)` sebelum dimasukkan ke struct DTO. Jangan return key mentah.

3. **Guard nil check**: jika field bersifat optional (`*string`), gunakan helper atau cek nil + empty sebelum memanggil `PublicURL`.

4. **Guard `://`**: jika nilai sudah mengandung `://` (full URL), lewati konversi — ini untuk backward-compatibility data lama yang sudah menyimpan full URL.

5. **Saat menyimpan ke domain** (create/update): simpan hanya key-nya, bukan full URL. Gunakan `fileUploader.KeyFromURL(rawURL)` atau `mediaUploadSvc.NormalizeKey(rawURL)` untuk mengekstrak key dari URL yang dikirim client.

### Contoh Alur Lengkap

```
Client kirim: image_url = "https://cdn.../images/uuid.jpg"
                ↓ usecase: mediaUploadSvc.NormalizeKey(url) → "/images/uuid.jpg"
Domain simpan: entity.ImageURL = "/images/uuid.jpg"  (key)
                ↓ saat baca & map ke DTO:
DTO kembalikan: image_url = fileUploader.PublicURL("/images/uuid.jpg")
              → "https://cdn.../images/uuid.jpg"
```

---

## 11. Handler Test: Wajib Ada, Wajib Dicek

### Aturan Utama

1. **Saat menambah atau mengubah handler** → cek apakah file `*_handler_test.go` untuk handler tersebut sudah ada di direktori yang sama.
   - Jika **belum ada** → buat file test sekaligus, ikuti pola di bawah.
   - Jika **sudah ada** → tambah/perbarui test yang relevan dengan perubahan.

2. **Saat mengerjakan endpoint baru** → test handler wajib dibuat sebelum dianggap selesai. Endpoint tanpa test = tidak selesai.

3. **Saat review atau debugging handler** → jalankan test yang ada sebelum dan sesudah perubahan.

### Lokasi Test

Tempatkan test handler di direktori yang sama dengan handler-nya, dengan package `_test` yang sesuai (mis. handler di package `mobile` → test di package `mobile_test`).

### Menjalankan Test

```bash
# Semua handler test pada satu grup
go test ./internal/interfaces/http/handler/<group>/... -count=1 -timeout 300s

# Filter per domain/kasus
go test ./internal/interfaces/http/handler/<group>/ -run "TestXxx" -v -timeout 120s
```

### Skenario Test Wajib per Domain

Setiap domain handler test **wajib** memiliki skenario:

- **Unauthenticated** (401) untuk semua protected endpoint
- **Invalid payload / validation error** (400 atau 422) untuk endpoint dengan body
- **Not found** (404) untuk endpoint yang fetch by ID
- **Forbidden / tidak memenuhi syarat akses** (403) untuk endpoint yang butuh entitlement/benefit tertentu
- **Success** (2xx) untuk happy path

### Infrastruktur Test

```
internal/testhelper/
  testserver.go   → TestServer, GET, POST, PUT, PATCH, DELETE, DELETEBody
  seed.go         → helper MustSeedXxx untuk data prasyarat (plan, role, dsb.)
  fixtures.go     → data fixture statis

internal/interfaces/http/handler/<group>/
  main_test.go        → TestMain, var testSrv
  testutil_test.go    → helper mustXxx (register user, login, seed data, dsb.)
```

### Nuance Penting (Lessons Learned)

Bagian ini adalah tempat mencatat kuirk/nuance spesifik proyek yang tidak terlihat dari membaca kode sekilas — misalnya kolom nullable yang butuh handling khusus, perilaku driver DB tertentu, konvensi penamaan tabel/kolom yang menyimpang dari asumsi umum, atau aturan validasi yang berbeda dari ekspektasi (mis. kode status binding Gin). Tambahkan catatan baru di sini setiap kali menemukan nuance semacam ini agar tidak perlu ditemukan ulang oleh sesi berikutnya.

---

## 12. Penamaan File Migrasi (Wajib Timestamp, Bukan Sequential)

File migrasi di `internal/migrations/` **wajib** dibuat dengan prefix timestamp UTC (`YYYYMMDDHHMMSS`), bukan nomor urut manual (`0026`, `0027`, dst).

**Cara membuat migrasi baru:**

```bash
make migrate-create NAME=nama_migrasi
```

Ini menghasilkan sepasang file, contoh:

```
internal/migrations/20260703031708_nama_migrasi.up.sql
internal/migrations/20260703031708_nama_migrasi.down.sql
```

**Alasan:** dengan penomoran sequential (`0025`, `0026`, ...), dua branch yang bercabang dari commit yang sama bisa menghasilkan file migrasi dengan nomor yang identik. Jika salah satu branch sudah di-deploy duluan, branch kedua bisa gagal total (`duplicate migration version` saat merge) atau — lebih berbahaya — migrasinya **diam-diam tidak pernah dijalankan** (golang-migrate menganggap versi tersebut sudah diterapkan karena DB sudah berada di versi yang sama). Timestamp per-detik membuat tabrakan semacam ini praktis mustahil.

**Aturan:**

- Jangan buat file migrasi manual dengan menyalin nomor urut dari file sebelumnya. Selalu pakai `make migrate-create NAME=...`.
- File migrasi lama dengan format sequential (bila ada) **tidak perlu diubah** — golang-migrate mem-parsing versi sebagai angka lalu mengurutkan secara numerik (bukan alfabetis), sehingga campuran format lama dan baru di folder yang sama tetap aman dan berjalan sesuai urutan.
- `NAME` pakai `snake_case` deskriptif, contoh: `add_outbox_correlation_id`, bukan singkatan ambigu.

---

## 13. Aturan Media Privat (Private Bucket)

§9 dan §10 di atas membahas alur media **public** (dibaca siapa saja via CDN URL statis). Section ini membahas alur **private** — untuk media sensitif yang aksesnya harus dibatasi dan time-limited.

### Kapan Pakai `PrivacyPrivate`

Gunakan `port.PrivacyPrivate` untuk media yang secara semantik sensitif dan tidak boleh diakses publik tanpa otorisasi, contoh: dokumen identitas/verifikasi, bukti pembayaran, lampiran internal admin. Audit modul yang masih memakai `PrivacyPublic` untuk jenis media semacam ini sebagai kandidat migrasi ke `PrivacyPrivate`.

### Dua Bucket Fisik Berbeda

Object storage punya dua bucket: `bucket` (public, anonymous-readable) dan `privateBucket` (tidak ada policy anonymous). Privacy ditentukan oleh bucket tujuan, bukan ACL per-objek. `RequestUpload`/`UploadFromReader`/`DeleteObject` semuanya menerima parameter `privacy PrivacyRule` untuk memilih bucket yang benar — **privacy yang dipakai saat upload harus konsisten dengan privacy yang dipakai saat delete**, karena disimpan per-record di kolom `media_uploads.privacy`.

### Kontrak `RequestUpload` untuk Private

```go
presignURL, key, publicURL, err := mediaUploadSvc.RequestUpload(ctx, objectName, contentType, expires, port.PrivacyPrivate)
```

Untuk `PrivacyPrivate`, **`publicURL` selalu `""`** — file tidak ada di bucket public, jadi nilai ini tidak berguna. Usecase presign **wajib membuang** nilai ini (`_` atau tidak dimasukkan ke DTO response field `public_url`). Jangan copy-paste pola domain public yang memakai `publicURL` di response.

### Resolve Akses — `AccessURL`

Untuk menghasilkan URL akses (dipanggil saat mapping domain object ke DTO baca), gunakan:

```go
url, err := mediaUploadSvc.AccessURL(ctx, prefixedValue, port.PrivacyPrivate, ttl)
if err != nil {
    url = "" // best-effort — satu media gagal presign tidak boleh menggagalkan seluruh response
}
```

- `PrivacyPublic` → delegasi ke `PublicURL(key)`, tidak pernah error.
- `PrivacyPrivate` → presigned GET URL time-limited. **Tidak ada fallback ke `PublicURL`** jika presign gagal — fallback semacam itu akan selalu 404 karena file private tidak ada di bucket public.

**Jangan** panggil `fileUploader.GeneratePresignedDownloadURL` langsung dari usecase — selalu lewat `mediaUploadSvc.AccessURL`, agar logic privacy-branching tidak terduplikasi di setiap domain.

### Konvensi TTL

TTL signed URL adalah **named constant per usecase** (bukan magic number inline), ditaruh di file usecase yang memakainya. Rekomendasi: 15 menit untuk dokumen yang direview manual oleh admin/superadmin (cukup lama untuk satu sesi review, cukup pendek untuk membatasi kebocoran link).

### `DeleteObject` Butuh Parameter Privacy

```go
DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
```

Kalau memanggil manual di luar jalur cleanup normal, pastikan privacy yang dipassing sama dengan privacy record aslinya — ambil dari `MediaUpload.Privacy` bila tersedia, jangan hardcode.

### Checklist Migrasi Domain Public → Private

Kalau ada domain existing yang mau dipindah dari `PrivacyPublic` ke `PrivacyPrivate`:

1. Ubah `RequestUpload(..., port.PrivacyPublic)` → `port.PrivacyPrivate` di usecase presign.
2. Buang penggunaan `publicURL` di DTO response presign — jangan diisi ke field `public_url`.
3. Ganti resolve URL saat baca (`resolveXxxImageURL` yang lama memanggil `fileUploader.PublicURL`) → `mediaUploadSvc.AccessURL(ctx, value, port.PrivacyPrivate, ttl)`.
4. Tentukan TTL yang sesuai konteks (named constant), jangan reuse konstanta domain lain begitu saja.
5. Jalankan `go build ./...` — perubahan privacy tidak mengubah signature apapun, jadi seharusnya tidak ada breaking call-site.
