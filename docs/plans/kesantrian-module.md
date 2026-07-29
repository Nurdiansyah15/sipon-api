# Kesantrian Module — Santri Entity, Dokumen, Admin Management & UI

Status: implemented (BE) / implemented (FE).
Scope: `sipon-api` (backend) + `sipon-ui` (frontend).

## Context

Sipon needs a "Kesantrian" module — a domain for managing santri (students) data,
documents, and the admin workflow for creating santri accounts and approving requests.

Key decisions from requirements:
- **Santri 1:1 User** — every santri has exactly one user account.
- **NIS as login identifier** — NIS (10 chars, pattern `1000[12][0-9]{5}`) becomes an
  additional `LoginIdentifierKind = "NIS"` alongside EMAIL/PHONE/USERNAME.
- **Admin-managed accounts** — admin creates santri via auto-generated password,
  displayed once (follows system-management pattern).
- **Santri request flow** — existing users without a santri row can request; admin
  approves with NIS or rejects with notes.
- **Documents in private bucket** — all santri documents stored in MinIO private bucket,
  accessed via presigned GET URLs (TTL 15 min).
- **Profile update scope** — PUT /santri/profile only updates non-identifier fields
  (excludes nis, username, email, phone). Fullname is editable.

### NIS Format

```
1000  [12]  [0-9]{2}  [0-9]{3}
  │     │       │         └── 3 digit sekuensial (000-999)
  │     │       └── 2 digit (00-99)
  │     └── gender (1 = Laki-laki, 2 = Perempuan)
  └── prefix tetap
```

NIS value object: `domain/santri/valueobject/santri_vo.go` — validates regex, exposes
`.Gender()` method to auto-set santri `option` field on create.

---

## Backend plan (sipon-api) — Implemented

Follow `.claude/CLAUDE.md` conventions: one usecase per file, DDD layers, repository
interfaces in domain, postgres impl in infrastructure, mandatory handler tests.

### 1. Domain Layer

#### Entities
| Entity | File | Description |
|--------|------|-------------|
| `Santri` | `domain/santri/entity/santri.go` | 1:1 User. Fields from Blade template (data pribadi, kontak, kependudukan, pendidikan, keluarga, pondok sebelumnya). NIS field using `*valueobject.NIS`. |
| `SantriDokumen` | `domain/santri/entity/santri_dokumen.go` | 1:N Santri. Kind (surat_pernyataan/ktp/kk/mutasi/pembayaran), status (pending/verified/rejected), key (private bucket). Methods: `Verify()`, `Reject()`, `SoftDelete()`. |
| `SantriRequest` | `domain/santri/entity/santri_request.go` | User request to become santri. Status (pending/approved/rejected), reviewed_by, reviewed_at. Methods: `Approve(nis)`, `Reject(notes)`. |

#### Value Objects
| VO | File | Description |
|----|------|-------------|
| `NIS` | `domain/santri/valueobject/santri_vo.go` | Regex `^1000[12][0-9]{5}$`, `.Gender()` extracts char index 4. |

#### Constants
| File | Content |
|------|---------|
| `domain/santri/constant/santri_constant.go` | `DokumenKind` enum, `DokumenStatus` enum, `SantriRequestStatus` enum, error codes (CodeSantriNotFound, CodeDokumenNotFound, CodeSantriRequestNotFound, CodeInvalidNISFormat, etc.) |
| `domain/user/constant/user_constant.go` | +`LoginIdentifierNIS LoginIdentifierKind = "NIS"`, +`CodeInvalidNISFormat` |

#### Repository Interfaces
| Interface | File | Methods |
|-----------|------|---------|
| `SantriRepository` | `domain/santri/repository/interfaces.go` | Save, Update, FindByID, FindByUserID, FindByNIS, List (paginated) |
| `SantriDokumenRepository` | same file | Save, Update, FindByID, FindBySantriID, FindBySantriIDAndKind, Delete |
| `SantriRequestRepository` | same file | Save, Update, FindByID, FindPendingByUserID, FindByStatus, List (paginated) |

### 2. User Domain Changes

**NIS as login identifier:**
- `user/valueobject/user_vo.go`: `NewLoginIdentifier()` now detects NIS pattern after email/phone and before username fallback.
- `user/entity/login_identity.go`: `normalizeLoginIdentityValue()` handles `LoginIdentifierNIS` case.
- Migration: `user_identities` CHECK constraint updated to include `'NIS'`.

**Login fallback:**
- `usecase/auth/login.go`: if `FindByLoginIdentifier(kind=NIS)` fails, retries with `FindByIdentity(kind=USERNAME)` for backward compatibility.

### 3. Infrastructure (Persistence)

| File | Implements |
|------|------------|
| `persistence/postgres_santri_repository.go` | `SantriRepository` — full CRUD with NIS column, `List` with pagination (ILIKE search on NIS) |
| `persistence/postgres_santri_dokumen_repository.go` | `SantriDokumenRepository` |
| `persistence/postgres_santri_request_repository.go` | `SantriRequestRepository` — `List` with status filter, unique constraint on pending requests per user |

**Port extension:**
- `app/port/file_uploader.go`: +`GeneratePresignedDownloadURL(ctx, key, privacy, expiry)` — needed for private bucket access.
- `infrastructure/external/minio/client.go`: implemented via `PresignedGetObject`.
- `infrastructure/external/minio/noop.go`: noop variant.

**Media object paths:**
- `app/service/media/object_path.go`: +`ObjectPathSantriDokumen = "/santri/dokumen/"`, +`SantriDokumenPresignUploadExpiry = 10min`, +`SantriDokumenAccessTTL = 15min`.

### 4. Application Layer (Usecases)

| Usecase | File | Role | Description |
|---------|------|------|-------------|
| `GetSantri` | `get_santri.go` | any (JWT) | Fetch santri + user data (fullname, email, phone, avatar). Returns NIS from santri entity. |
| `UpdateSantri` | `update_santri.go` | any (JWT) | Update non-identifier santri fields + fullname via UserRepo. NIS immutable after create. |
| `CreateSantri` | `create_santri.go` | manage_users | Admin creates User + Santri in one flow. Auto-generates password, email, username from NIS. Gender auto-set from NIS.Gender(). Returns generated password. |
| `ListSantri` | `list_santri.go` | manage_users | Paginated listing with NIS search. Enriches with User data (fullname, username, email, status). |
| `RequestSantri` | `request_santri.go` | any (JWT) | User without santri row creates a pending SantriRequest. Guards against duplicate requests. |
| `ListSantriRequests` | `list_santri_requests.go` | manage_users | Paginated listing with status filter. Enriches with User data. |
| `ApproveSantriRequest` | `approve_santri_request.go` | manage_users | Approves request with NIS. Creates Santri row + adds NIS LoginIdentity to User. |
| `RejectSantriRequest` | `reject_santri_request.go` | manage_users | Rejects request with optional notes. |
| `DokumenPresign` | `dokumen_request_upload.go` | any (JWT+santri) | Generates presigned PUT URL for private bucket upload. Validates content type + kind. |
| `DokumenConfirm` | `dokumen_confirm_upload.go` | any (JWT+santri) | Creates SantriDokumen row after successful upload. Confirms media key. |
| `DokumenList` | `dokumen_list.go` | any (JWT+santri) | Lists documents with optional kind filter. |
| `DokumenAccess` | `dokumen_access.go` | any (JWT+santri) | Generates presigned GET URL (private bucket, TTL 15min). Ownership check. |
| `DokumenDelete` | `dokumen_delete.go` | any (JWT+santri) | Soft delete + remove from object storage. Ownership check. |
| `DokumenVerify` | `dokumen_admin_verify.go` | manage_users | Admin verifies document. |
| `DokumenReject` | `dokumen_admin_reject.go` | manage_users | Admin rejects document with optional notes. |

### 5. HTTP Routes

```
/api/v1/web/santri
├── GET  /profile                                    → GetSantri
├── PUT  /profile                                    → UpdateSantri
├── POST /request                                    → RequestSantri
├── POST /dokumen/presign                            → DokumenPresign
├── POST /dokumen/confirm                            → DokumenConfirm
├── GET  /dokumen                                    → DokumenList
├── GET  /dokumen/:id/access                         → DokumenAccess
├── DELETE /dokumen/:id                              → DokumenDelete
│
└── /admin (RequirePermission: manage_users)
    ├── GET  /                                       → ListSantri
    ├── POST /                                       → CreateSantri
    ├── GET  /requests                               → ListSantriRequests
    ├── POST /requests/:id/approve                   → ApproveSantriRequest
    ├── POST /requests/:id/reject                    → RejectSantriRequest
    ├── POST /verify/:id                             → DokumenVerify
    └── POST /reject/:id                             → DokumenReject
```

### 6. Migrations

| Migration | Content |
|-----------|---------|
| `20260729032931_create_santri_tables` | Table `santri` (all fields flat, 53 cols), table `santri_dokumen` (kind, key, status, metadata) |
| `20260729065536_add_santri_nis_and_requests` | ALTER santri add `nis VARCHAR(10) UNIQUE`, update `user_identities` CHECK for NIS kind, create `santri_requests` table |

---

## Frontend plan (sipon-ui) — Implemented

### 1. New Shared Types

| File | Content |
|------|---------|
| `shared/types/Santri.ts` | `SantriProfile`, `UpdateSantriRequest`, `CreateSantriRequest/Response`, `ListSantriItem`, `ListSantriQuery`, `SantriRequestItem`, `ApproveSantriRequestReq`, `RejectSantriRequestReq`, `RequestSantriRes` |
| `shared/types/SantriDokumen.ts` | `DokumenPresignReq/Res`, `DokumenConfirmReq/Res`, `DokumenItem`, `DokumenAccessRes` |

### 2. Pinia Stores

**`stores/santri.ts`:**
| Action | API | Description |
|--------|-----|-------------|
| `fetchProfile()` | `GET /santri/profile` | Load santri profile (santri + user data) |
| `updateProfile(data)` | `PUT /santri/profile` | Update non-identifier fields |
| `requestSantri()` | `POST /santri/request` | Submit request to become santri |
| `fetchDokumenList(kind?)` | `GET /santri/dokumen` | Load document list, optional kind filter |
| `requestDokumenPresign(payload)` | `POST /santri/dokumen/presign` | Get presigned PUT URL |
| `confirmDokumen(payload)` | `POST /santri/dokumen/confirm` | Confirm upload |
| `getDokumenAccess(id)` | `GET /santri/dokumen/:id/access` | Get presigned GET URL |
| `deleteDokumen(id)` | `DELETE /santri/dokumen/:id` | Soft delete document |

**`stores/pengelolaSantri.ts`:**
| Action | API | Description |
|--------|-----|-------------|
| `listSantri(query)` | `GET /santri/admin` | Paginated santri list with NIS search |
| `createSantri(nis)` | `POST /santri/admin` | Create santri, stores oneTimePassword |
| `listRequests(query)` | `GET /santri/admin/requests` | Paginated request list with status filter |
| `approveRequest(id, nis)` | `POST /santri/admin/requests/:id/approve` | Approve with NIS |
| `rejectRequest(id, notes)` | `POST /santri/admin/requests/:id/reject` | Reject with optional notes |
| `verifyDokumen(id)` | `POST /santri/admin/verify/:id` | Verify document |
| `rejectDokumen(id, notes)` | `POST /santri/admin/reject/:id` | Reject document |

### 3. Routes & Pages

#### Santri Profile (layout: `default`)
| Route | File | Auth | Description |
|-------|------|------|-------------|
| `/santri` | `pages/santri/index.vue` | JWT | Landing: fetches profile via GET /santri/profile. If 404 → show `RequestCard`. If success → show `ProfileCard` + grid links to Profile and Dokumen. |
| `/santri/profile` | `pages/santri/profile.vue` | JWT + santri | Tabbed edit form (6 tabs). Redirects to /santri if no santri row. |
| `/santri/dokumen` | `pages/santri/dokumen.vue` | JWT + santri | Upload (presign→PUT→confirm) + document list with view/delete. |

#### Pengelola Santri (layout: `pengelola-santri`)
| Route | File | Auth | Description |
|-------|------|------|-------------|
| `/pengelola-santri` | `pages/pengelola-santri/index.vue` | JWT + manage_users | Admin dashboard: stats cards (total santri, pending requests), navigation cards. |
| `/pengelola-santri/santri` | `pages/pengelola-santri/santri/index.vue` | JWT + manage_users | `UTable` with NIS search, pagination. Create button → `CreateSantriModal`. |
| `/pengelola-santri/requests` | `pages/pengelola-santri/requests/index.vue` | JWT + manage_users | `UTable` with status filter, pagination. Approve/Reject modals. |

### 4. Layout: `pengelola-santri`

New layout mirroring `system-admin`:
- `layouts/pengelola-santri.vue` — Wraps `AppPengelolaSantriNavbar` + `<slot>` + `AppFooter` + `AppPengelolaSantriMobileNav`.
- `AppPengelolaSantriNavbar.vue` — Desktop nav: "Dasbor", "Data Santri", "Request", + kebab nav to dashboard.
- `AppPengelolaSantriMobileNav.vue` — Mobile bottom nav: same items.

### 5. Components

#### Santri Components (`components/santri/`)
| Auto-import Name | File | Purpose |
|------------------|------|---------|
| `SantriProfileCard` | `ProfileCard.vue` | Avatar + NIS + name display card |
| `SantriRequestCard` | `RequestCard.vue` | CTA card: "Ajukan Sebagai Santri" button |
| `SantriProfileForm` | `ProfileForm.vue` | Tabbed form (UTabs) with 6 sections + submit |
| `SantriDataPribadiSection` | `DataPribadiSection.vue` | Form: nickname, hobby, purpose, motivation_entry, pob, dob, blood |
| `SantriKontakSection` | `KontakSection.vue` | Form: address, sub_district, district, province, postal_code |
| `SantriKependudukanSection` | `KependudukanSection.vue` | Form: nik, no_kk, nisn, no_kip, no_kks, no_pkh |
| `SantriPendidikanSection` | `PendidikanSection.vue` | Form: workplace, department |
| `SantriKeluargaSection` | `KeluargaSection.vue` | Form: home_status + Ayah/Ibu/Wali sub-sections |
| `SantriPondokSection` | `PondokSection.vue` | Form: previous_pondok_* |
| `SantriDokumenUploadCard` | `DokumenUploadCard.vue` | Upload widget: select kind + file → presign → PUT → confirm |
| `SantriDokumenListTable` | `DokumenListTable.vue` | List with kind, filename, status badge, view/delete actions |
| `SantriDokumenViewModal` | `DokumenViewModal.vue` | UModal with iframe using presigned GET URL |

#### Pengelola Santri Components (`components/pengelola-santri/`)
| Auto-import Name | File | Purpose |
|------------------|------|---------|
| `PengelolaSantriCreateSantriModal` | `CreateSantriModal.vue` | UModal: input NIS (Zod validated) → submit → show generated password |
| `PengelolaSantriApproveRequestModal` | `ApproveRequestModal.vue` | UModal: input NIS → approve |
| `PengelolaSantriRejectRequestModal` | `RejectRequestModal.vue` | UModal: optional notes → reject |

### 6. Edited Existing Components

| File | Change |
|------|--------|
| `FeatureModuleGrid.vue` | "Kesantrian" card now has `to: '/santri'` — clickable |
| `AppSystemAdminNavbar.vue` | +"Pengelola Santri" nav item (gated: `can('manage_users')`) → `/pengelola-santri` |

### 7. Document Upload Flow (Presign Pattern)

```
Client                    API                        MinIO
  │                        │                          │
  │-- POST /presign       │                          │
  │   {content_type, kind}│                          │
  │                        │-- RequestUpload -------> │  (PrivacyPrivate)
  │<-- presign_url, key    │                          │
  │                        │                          │
  │-- PUT presign_url      │                         │  (file binary, direct)
  │                        │                          │
  │-- POST /confirm       │                          │
  │   {kind, key, meta}   │                          │
  │                        │-- Save SantriDokumen --> │
  │<-- 201                 │                          │
  │                        │                          │
  │  Doc list refresh      │                          │
  │                        │                          │
  │-- GET /:id/access     │                          │
  │                        │-- PresignedGetObject --> │  (TTL 15min)
  │<-- access_url          │                          │
```

### 8. Navigation & Permission Gating

- **Dashboard → Kesantrian**: All logged-in users see the clickable card at `/santri`.
- **System Admin → Pengelola Santri**: Users with `manage_users` permission see nav item in AppSystemAdminNavbar.
- **Pengelola Santri layout**: All pages gated by JWT; admin-specific pages further gated by `manage_users` at the backend route level.
- **Component-level gating**: `SantriRequestCard` only shows when no santri row exists (based on API 404).
- **Dokumen access**: Only owned documents are accessible/resolvable — backend checks `santri_id` ownership.

### 9. Tech Stack Notes (Nuxt UI v4 conventions)

- **UTable**: uses `:data`, `:columns` with `{ accessorKey, header }` or `{ id, header }` for custom columns. Cell templates via `#column_id-cell="{ row }"` where `row.original.X` accesses data.
- **Auto-import naming**: Subdirectory components auto-import with directory prefix. E.g., `components/santri/ProfileCard.vue` → `SantriProfileCard`.
- **Modals**: Use `:open` + `@close` pattern with `<UModal>`, not `v-model:open`.
- **Forms**: Zod + `UForm` + `reactive<Partial<Schema>>()`. Always `.refine()` for cross-field validations.
- **Zod NIS validation**: `z.string().length(10).regex(/^1000[12]\d{5}$/, 'Format NIS tidak valid: 1000[12]xxxxx')`.
- **Presign upload**: File → fetch presign endpoint → `fetch(presign_url, { method: 'PUT', body: file })` → confirm endpoint.
- **Private document viewing**: GET `/dokumen/:id/access` → render presigned URL in `<iframe>`.
