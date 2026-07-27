# User Profile Feature Spec

**Status:** Implemented (BE: consolidated endpoint + change-password error mapping)
**Scope:** `sipon-api` (BE) + `sipon-ui` (FE)
**Page route:** `/profile` (layout: `default`)
**Menu:** "My Profile" di `AppUserMenu.vue` → `navigateTo('/profile')`

---

## 1. API Endpoints

### Existing (sudah tersedia, tinggal dikonsumsi)

| Method | Endpoint | Response | Untuk Tab |
|--------|----------|----------|-----------|
| `GET` | `/api/v1/web/auth/me` | `UserMe` — id, username, fullname, email, is_email_verified, phone, is_phone_verified, status, created_at, has_password | Informasi Akun |
| `GET` | `/api/v1/auth/session` | `SessionData` — user (summary) + roles[] + permissions[] | Roles & Permissions |
| `POST` | `/api/v1/web/auth/change-password` | body: `{ current_password, new_password, new_password_confirmation }` | Keamanan |
| `POST` | `/api/v1/web/auth/set-password` | body: `{ new_password, new_password_confirmation }` | Keamanan (jika has_password=false) |

### Baru / Enhanced (sudah diimplementasikan)

**Consolidated Profile Endpoint** (merge `/me` + `/session`)

`GET /api/v1/web/auth/profile`

```json
{
  "id": "uuid",
  "username": "johndoe",
  "fullname": "John Doe",
  "email": "john@example.com",
  "is_email_verified": true,
  "phone": "08123456789",
  "is_phone_verified": true,
  "status": "active",
  "has_password": true,
  "created_at": "2026-07-27T00:00:00Z",
  "roles": [
    { "name": "admin", "role_type": "system", "scope_type": "global", "scope_id": null }
  ],
  "permissions": [
    { "key": "manage_users", "scope": "global" },
    { "key": "assign_role", "scope": "global" }
  ]
}
```

> Opsi B (2 endpoint terpisah) sudah digantikan oleh consolidated endpoint di atas. FE cukup panggil satu endpoint `/auth/profile`.

---

## 2. Halaman & Tab Structure

```
┌─────────────────────────────────────────────────┐
│  Header: "Profil Saya"                          │
├─────────────────────────────────────────────────┤
│  [ Informasi Akun ] [ Roles & Permissions ] [ Keamanan ] │
├─────────────────────────────────────────────────┤
│                                                   │
│  (konten sesuai tab aktif)                        │
└─────────────────────────────────────────────────┘
```

### Tab 1: Informasi Akun

Menampilkan data user dalam bentuk **read-only card/info list** (belum termasuk edit profile di fase ini).

Field yang ditampilkan:

| Label | Source field | Tampilan |
|-------|-------------|----------|
| Username | `user.username` | Text |
| Nama Lengkap | `user.fullname` | Text atau "—" jika null |
| Email | `user.email` | Text + badge "Terverifikasi" (hijau) / "Belum Verifikasi" (kuning) |
| No. Telepon | `user.phone` | Text atau "—" jika null + badge verifikasi |
| Status Akun | `user.status` | Badge: "Aktif" (hijau) / "Diblokir" (merah) |
| Anggota Sejak | `user.created_at` | Format tanggal Indonesia |
| Kata Sandi | `user.has_password` | Badge "Sudah diatur" (hijau) / "Belum diatur" (kuning) |

Data source: `GET /api/v1/web/auth/profile` (consolidated, mencakup data `/me` + `/session`).

### Tab 2: Roles & Permissions

Dua sub-seksi:

**a. Roles yang dimiliki:**
Daftar role cards/badges, masing-masing menampilkan:
- Nama role (`display_name`)
- Tipe role (`role_type`: "System" vs "Custom" — badge)
- Scope (`scope_type` + `scope_id` jika ada)

**b. Permissions efektif (aggregated):**
Tabel atau grid semua permission keys yang dimiliki user saat ini:

| Permission | Deskripsi |
|-----------|-----------|
| `manage_users` | Mengelola akun user... |
| `assign_role` | Menetapkan peran ke user... |

Setiap permission juga bisa diberi badge asal role (misal: "dari role: admin").

Data source: `GET /api/v1/auth/session`.

### Tab 3: Keamanan

Dua kemungkinan form (hanya satu yang aktif tergantung kondisi):

**a. Set Password** (jika `user.has_password === false`)
- Form: `New Password` + `Confirm New Password`
- Validasi: min 8 chars, 1 uppercase, 1 digit, confirm match
- Submit → `POST /api/v1/web/auth/set-password`
- Success: toast + refresh `has_password` jadi `true`, form berganti ke "Ubah Password"

**b. Ubah Password** (jika `user.has_password === true`)
- Form: `Current Password` + `New Password` + `Confirm New Password`
- Validasi: min 8 chars, 1 uppercase, 1 digit, confirm match, current password wajib diisi
- Submit → `POST /api/v1/web/auth/change-password`
- Success: toast

---

## 3. Frontend Implementation Details

**File baru:**
- `app/pages/profile/index.vue` — halaman utama, layout `default`

**Komponen (opsional, bisa inline):**
- `app/components/profile/AccountInfoPanel.vue`
- `app/components/profile/RolesPermissionsPanel.vue`
- `app/components/profile/ChangePasswordForm.vue`
- `app/components/profile/SetPasswordForm.vue`

**State management:**
- Tidak perlu store baru. Data dari `useAuthStore` sudah cukup:
  - `authStore.user` → profile info
  - `authStore.roles` → roles
  - `authStore.permissions` → permissions
- Panggil consolidated endpoint `GET /api/v1/web/auth/profile` via `useApi()` di `onMounted` page

**Form pattern:**
- Zod schema + `UForm` + `UFormField` + `UInput` (existing pattern)
- Submit via `useApi()` atau store action (perlu ditambahkan di `authStore` untuk change/set password)

**Tabs pattern:**
- Gunakan `UTabs` dari `@nuxt/ui` v4
- Tab items: `[{ label: 'Informasi Akun', icon: 'i-lucide-user' }, { label: 'Roles & Permissions', icon: 'i-lucide-shield' }, { label: 'Keamanan', icon: 'i-lucide-lock' }]`
- Gunakan `v-model` untuk tab index, konten di-render conditional

**Permission/role gate di tab Roles & Permissions:**
- Tidak perlu gate khusus — endpoint session sudah mengembalikan data user itu sendiri
- Tapi untuk edukasi, bisa tampilkan "Anda tidak memiliki roles/permissions" jika array kosong

---

## 4. Pekerjaan (Tasks)

### BE (sipon-api)

| # | Task | Prioritas | Status |
|---|------|-----------|--------|
| 1 | Buat consolidated endpoint `GET /api/v1/web/auth/profile` yang return `UserMe` + `SessionData` dalam satu response | Medium | ✅ Done |
| 2 | Pastikan `change-password` dan `set-password` sudah berfungsi dengan validasi yang tepat | High | ✅ Done |
| 3 | Tambahkan error code mapping untuk password flows (wrong current password → 422, etc.) | High | ✅ Done |

### FE (sipon-ui)

| # | Task | Prioritas |
|---|------|-----------|
| 1 | Buat `app/pages/profile/index.vue` dengan layout default | High |
| 2 | Implement UTabs dengan 3 tab | High |
| 3 | Tab "Informasi Akun": consume `authStore.user`, tampilkan data read-only | High |
| 4 | Tab "Roles & Permissions": consume `authStore.roles` + `authStore.permissions` | High |
| 5 | Tab "Keamanan": Set Password form (jika `has_password=false`) | High |
| 6 | Tab "Keamanan": Change Password form (jika `has_password=true`) | High |
| 7 | Update `AppUserMenu.vue`: arahkan "My Profile" ke `/profile` | High |
| 8 | Tambahkan action `changePassword` / `setPassword` di `authStore` | High |
| 9 | Handle loading state, error toast, success toast | Medium |

---

## 5. Catatan / Keputusan

1. **Tab-driven**: Tidak ada nested routing — semua tab di halaman yang sama, konten di-render dengan `v-if`/`v-show`.
2. **Read-only profile**: Fase ini tidak termasuk form edit profile (nama, email, phone). Murni menampilkan data.
3. **Setelah sukses set-password**: Langsung refresh `authStore.user` agar badge/tab berubah ke "Ubah Password". Bisa pakai `fetchMe()` lagi.
