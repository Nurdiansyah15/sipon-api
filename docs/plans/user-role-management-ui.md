# User Role Management — Perbaikan Halaman Kelola User

Status: planned.
Scope: `sipon-api` (this repo) + `sipon-ui` (sibling frontend).

## Context

Halaman "Kelola User" (`/system-admin/users`) saat ini:

- Tabel user hanya menampilkan username, email, status, dan tanggal dibuat — **tidak ada kolom
  "Roles"** yang menunjukkan role apa saja yang dimiliki tiap user.
- Modal "Kelola Role" (`AssignRoleModal.vue`) adalah form assign-role semata — **tidak menampilkan
  role user saat ini** dan **tidak bisa menghapus role**. Admin harus menduga-duga role yang sudah
  dimiliki user, atau mengecek via endpoint/user detail secara manual.

Tujuan perbaikan ini:

1. Admin **melihat role tiap user langsung di tabel list** (sebagai badge/chip).
2. Modal "Kelola Role" menampilkan **daftar role user saat ini** dengan tombol hapus per role.
3. Modal yang sama tetap memiliki **form tambah role** untuk assign baru.
4. Setiap aksi hapus/tambah langsung tercermin di tabel tanpa reload halaman penuh.

### Current state

**Backend**:
- `GET /web/users` (list) sengaja **tidak mengisi field `Roles`** untuk menghindari N+1
  (lihat `list_users.go:52`).
- `GET /web/users/:user_id` sudah mengembalikan `roles` (via `ListActiveRoleSummariesByUserID`).
- Endpoint assign/remove role sudah tersedia: `POST /user-roles`, `DELETE /user-roles/:id`,
  `GET /user-roles?user_id=`.
- Tidak ada batch query untuk mengambil roles banyak user sekaligus.

**Frontend**:
- Tabel user (`users/index.vue`) tidak memiliki kolom "Roles".
- `AssignRoleModal.vue` adalah form satu arah (assign only), menutup dirinya setelah sukses tanpa
  menampilkan state role terkini.
- Store `rolePermission.ts` sudah punya `fetchUserRoles`, `assignUserRole`, `deleteUserRole` —
  siap dipakai tanpa perubahan.
- Type `UserManagementItem.roles` sudah didefinisikan sebagai `UserRoleSummary[]` di shared types.

---

## Backend plan (sipon-api)

### 1. Batch query roles untuk user list

**Masalah**: `ListUsersUseCase` tidak mengisi `Roles` karena khawatir N+1 — satu query ke
`ListActiveRoleSummariesByUserID` per user. Solusinya adalah **satu batch query** untuk semua user
sekaligus (`WHERE user_id IN (...)`), lalu mapping hasilnya ke masing-masing response item.

**1.1 Port interface** — `internal/app/port/user_query_model.go`

Tambahkan method ke interface `UserQueryReadModel`:

```go
ListActiveRoleSummariesByUserIDs(ctx context.Context, userIDs []string) (map[string][]UserRoleSummaryReadItem, error)
```

Mengembalikan `map[userID][]summary` agar usecase tinggal lookup O(1) per user.

**1.2 Persistence implementation** — `internal/infrastructure/persistence/postgres_user_query.go`

Implementasi dengan SQL `WHERE ur.user_id IN (...)` menggunakan dynamic placeholders (`database/sql`
tidak support array parameter seperti pgx):

```sql
SELECT ur.id, ur.role_id, r.name, ur.scope_type, ur.scope_id, ur.is_active, ur.user_id
FROM user_roles ur
INNER JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id IN ($1, $2, ...)
  AND ur.is_active = TRUE
  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
ORDER BY ur.assigned_at DESC
```

Hasil di-scan ke `map[string][]UserRoleSummaryReadItem`.

**1.3 Usecase** — `internal/app/usecase/usermanagement/list_users.go`

Setelah memanggil `readModel.ListUsers()` dan membangun `items []UserManagementResponse`:

1. Kumpulkan semua `userID` dari `items`.
2. Jika `len(userIDs) > 0`, panggil `readModel.ListActiveRoleSummariesByUserIDs(ctx, userIDs)`.
3. Iterasi `items`, lookup `map[item.ID]`, isi `item.Roles`.

### 2. Tidak ada API route baru

Semua endpoint yang diperlukan sudah ada:
- `GET /web/users` — setelah fix di atas, kini mengembalikan `roles`.
- `GET /web/role-permission/user-roles?user_id=` — untuk fetch detail role user saat modal dibuka.
- `POST /web/role-permission/user-roles` — assign role.
- `DELETE /web/role-permission/user-roles/:user_role_id` — remove role.

---

## Frontend plan (sipon-ui)

### 1. Kolom "Roles" di tabel user

**File**: `app/pages/system-admin/users/index.vue`

- Tambah kolom `{ accessorKey: 'roles', header: 'Roles' }` ke `columns`.
- Template slot `#roles-cell` merender `UBadge` untuk setiap role di `row.original.roles`,
  menampilkan `role_name`. Jika `roles` kosong/null, tampilkan "-" atau badge "No roles".
- Warna badge: `system` role pakai `neutral`, `custom` role pakai `teal` (informasi `role_type`
  tidak ada di `UserRoleSummary` — fallback semua pakai `neutral`).

### 2. Redesign modal "Kelola Role"

**File**: `app/components/system-admin/AssignRoleModal.vue`

Refactor total menjadi modal dua-bagian:

**Bagian atas: Daftar role saat ini**

- Saat modal terbuka (`watch open`), fetch `fetchUserRoles({ user_id: targetUserId, limit: 50 })`.
- Tampilkan sebagai list item, masing-masing menampilkan:
  - Nama role (`role.display_name`)
  - Scope type (badge)
  - Status aktif/nonaktif
  - Tombol hapus (ikon trash, warna error) — konfirmasi sebelum hapus (opsional, bisa langsung).
- Hapus: panggil `deleteUserRole(assignment.id)`, toast sukses, refresh list roles lokal (fetch
  ulang atau splice dari array).

**Bagian bawah: Form tambah role**

- Form assign-role yang sudah ada, dipertahankan tapi dipindah ke bawah.
- Setelah sukses assign, refresh daftar role lokal (tidak perlu fetch ulang tabel user utama).
- Emit event `updated` (rename dari `assigned`) agar parent bisa refresh tabel.

**State lokal** (tidak perlu store baru):
- `currentRoles: UserRoleItem[]` — hasil fetch `fetchUserRoles`.
- `isLoadingRoles: boolean` — loading state untuk fetch roles.
- `isDeleting: string | null` — tracking id yang sedang dihapus untuk loading indicator.

### 3. Store — tidak ada perubahan

`app/stores/rolePermission.ts` sudah memiliki semua method yang diperlukan. Method `fetchUserRoles`
mengisi `state.userRoles` — tapi karena modal bisa dibuka untuk user berbeda, lebih baik fetch
langsung via api composable di dalam modal (tidak mengandalkan shared store state) atau gunakan
local return value dari store action (saat ini `fetchUserRoles` tidak return apa-apa). Solusi:
tambahkan return value `Promise<UserRoleItem[]>` ke `fetchUserRoles` agar modal bisa menggunakannya
secara lokal tanpa membaca `store.userRoles`.

### 4. Types — tidak ada perubahan

- `UserRoleSummary` sudah ada di `shared/types/UserManagement.ts`.
- `UserRoleItem` sudah ada di `shared/types/RolePermission.ts`.
- `UserManagementItem.roles` sudah ada.

---

## Verification

- Backend: `go test ./internal/app/usecase/usermanagement/... -count=1 -timeout 120s`.
- Frontend: `bun dev`, login sebagai superadmin/admin:
  - Buka `/system-admin/users` — pastikan kolom "Roles" menampilkan badge role tiap user.
  - Klik "Kelola Role" pada user yang sudah punya 2+ role — pastikan semua role tampil.
  - Hapus satu role — pastikan role hilang dari daftar, toast sukses.
  - Tambah role baru — pastikan muncul di daftar.
  - Tutup modal — pastikan tabel user utama ikut ter-refresh (kolom Roles terupdate).
  - Coba assign role yang sudah dimiliki user — pastikan API return idempotent (existing
    assignment dikembalikan, bukan error).

## Critical files

- `internal/app/port/user_query_model.go` — tambah `ListActiveRoleSummariesByUserIDs`.
- `internal/infrastructure/persistence/postgres_user_query.go` — implementasi batch query.
- `internal/app/usecase/usermanagement/list_users.go` — panggil batch query + isi Roles.
- `../sipon-ui/app/pages/system-admin/users/index.vue` — tambah kolom Roles.
- `../sipon-ui/app/components/system-admin/AssignRoleModal.vue` — redesign penuh.
- `../sipon-ui/app/stores/rolePermission.ts` — tambah return value ke `fetchUserRoles`.
