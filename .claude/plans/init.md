# Init From Another Project# Plan: Bootstrap sipon-api dari k-forum-api (infra + auth/authorization)

## Context

`sipon-api` saat ini benar-benar kosong (tidak ada `go.mod`, tidak ada kode — hanya `.claude/CLAUDE.md` berisi panduan arsitektur DDD generik). Kita akan membangun fondasi proyek ini dengan mem-porting bagian-bagian tertentu dari `/home/nurdiansyah/Desktop/k-forum-api` (branch `development`, sudah checked out — dibaca read-only, tidak pernah checkout/switch branch), lalu menyesuaikan seluruh identitas proyek dari "k-forum"/"K-Forum" menjadi "sipon"/"Sipon".

Scope disepakati bersama user:
- **Infra**: arsitektur Docker, tooling migrasi database, tooling seeder database, tooling Swagger, Makefile.
- **Fitur**: User Auth + Authorization (RBAC) — **local-only** (tanpa Google/Apple login, tanpa seamless mobile↔web SSO).
- **Docker**: minimal — hanya `postgres` + `redis` (redis dipakai untuk session revocation store + principal cache). Tidak menyertakan rabbitmq/opensearch/minio/observability/deeplink.
- **RBAC**: disederhanakan ke scope `global` saja. Kolom `scope_type` (global/region/community) tetap dipertahankan di skema untuk fleksibilitas masa depan, tapi hanya role `usergod`, `superadmin`, `admin`, `member` yang di-seed (drop `guest`, `leader`, `moderator`, `community_member`). Tidak ada konsep subscription/plan/benefit — `RequireBenefit` middleware dan field `Plan`/`Benefits`/`AdminScopes` pada `Principal` **tidak diikutkan**.
- Hanya HTTP surface **web** yang diporting dulu (handler `mobile` di-skip untuk fase ini).

Hasil riset menyeluruh (docker-compose, Dockerfile, Makefile, migrasi, seeder, domain user/role, usecase auth, principal builder, middleware, router, testhelper) sudah dilakukan terhadap source. Semua path di bawah relatif ke masing-masing repo root kecuali disebutkan lain.

---

## Keputusan & Penyederhanaan Kunci

1. **Module path baru**: `sipon-api` (mirror 1:1 dari `k-forum-api`). Semua import internal → `sipon-api/internal/...`.
2. **Bootstrap seed user**: email `usergod@sipon.dev` (ganti dari `usergod@k-forum.dev`), dipakai di `internal/seeders/user_seeder.go` dan query `lookupSuperadminID` di `internal/seeders/seeder.go`.
3. **Docker naming**: network `sipon-net`, container `sipon-api_postgres`/`sipon-api_redis`/dst. `.env.example` default `DB_USER=sipon`, `DB_NAME=sipon`.
4. **Swagger**: `@title Sipon API`, `@description API development docs untuk Sipon.`
5. **Kolom/field yang di-drop** karena hanya dipakai fitur di luar scope: `users.is_verified` (verification-badge module, belum diporting), `credentials.apple_refresh_token`, `Avatar`/`AppleID`/`PendingAcceptances` pada `UserMe`/`SessionUser`/`LoginResponse` (butuh domain `user_profile`/legal yang belum ada).
6. **Outbox/event publishing di-drop** dari `register.go`/`login.go` — tidak ada worker/RabbitMQ untuk mengonsumsinya di scope minimal ini.
7. **i18n bundle di-drop** dari `httperror` — `Handle` langsung pakai message/kode error, tanpa translasi locale (bisa ditambah lagi nanti tanpa mengubah call-site).
8. **Permission catalog diperkecil**: hanya `manage_system_settings`, `assign_role`, `manage_users` (39 dari 42 permission asli berhubungan dengan domain yang belum diporting seperti news/community/qna/dll).
9. **`ListUsersUseCase`, Google/Apple login, seamless token usecases/handlers** — tidak diporting sama sekali.

Ringkasan lengkap rename "K-Forum → Sipon" ada di bagian akhir plan ini.

---

## Fase Implementasi

Kerjakan berurutan; setiap fase harus `go build ./...` sukses (begitu ada kode) sebelum lanjut ke fase berikutnya.

### Fase 1 — Go Module Bootstrap + Skeleton + Config
- `go.mod` (`module sipon-api`, `go 1.25`), dependencies: gin-gonic/gin, gin-contrib/cors, go-playground/validator/v10, golang-jwt/jwt/v5, golang-migrate/migrate/v4, google/uuid, jackc/pgx/v5, joho/godotenv, redis/go-redis/v9, stretchr/testify, swaggo/{files,gin-swagger,swag}, golang.org/x/crypto (bcrypt), testcontainers-go(+wait).
- `internal/config/config.go` — struct trimmed: `App`, `Database`, `JWT`, `SMTP` (from default `noreply@sipon.dev`), `Fonnte`, `Redis`, `RateLimit`, `Migration`. Drop Google/Apple/RabbitMQ/Minio/FCM/OpenSearch/Weather/ExchangeRate/Anthropic/Midtrans/AppBaseURL/BackofficeURL/SettingsEncryptionKey.
- `internal/logger/{logger.go,context.go,context_handler.go}` — port apa adanya (tidak ada branding).

### Fase 2 — Docker Architecture (minimal)
- `Dockerfile` — 4 stage (`builder`→`devtools`→`app`→`migrate`), **drop stage `worker`**.
- `docker-compose.dev.yml` — service: `postgres` (postgres:16-alpine, container `sipon-api_postgres`), `redis` (redis:7-alpine, container `sipon-api_redis`), `app` (golang:1.25-alpine, `go run ./cmd/app`, bind-mount, profile default), `migrate` & `seeder` & `devtools` (profile `tooling`). **Drop**: rabbitmq, opensearch(+dashboards), worker, minio(+init), deeplink. Network `sipon-net`, volumes `postgres_data`/`redis_data`/`go_mod_cache`/`go_build_cache`.
- `.env.example` — trimmed ke App/DB/Redis/Migration/JWT/SMTP/Fonnte/RateLimit saja.

### Fase 3 — Migration Tooling + Skema Auth/RBAC Konsolidasi
- `cmd/migrate/main.go` (port verbatim: up|down|fresh|version|force), `internal/migrations/embed.go` (`//go:embed *.sql`).
- **Satu migrasi baru** dibuat via `make migrate-create NAME=create_auth_rbac_tables` (timestamp asli di-generate saat implementasi), menggabungkan skema dari source `0001_create_auth_tables` + `0022_add_account_lockout`, dengan penyederhanaan:
  - `users`: id, username, fullname, email, phone, status(ACTIVE/BANNED), `failed_login_attempts`, `locked_until`, timestamps. Tanpa `is_verified`.
  - `credentials`: `type CHECK IN ('LOCAL')` saja (bukan +GOOGLE), tanpa `apple_refresh_token`.
  - `user_identities`: `kind CHECK IN ('EMAIL','PHONE','USERNAME')` saja (bukan +GOOGLE).
  - `roles`, `permissions`, `role_permissions`, `user_roles`: identik dengan source, **`scope_type` tetap 3 nilai** (global/region/community) untuk ekstensibilitas.
  - `verification_codes`: purpose tanpa `ACCOUNT_DELETION`.
  - Shared trigger `set_updated_at_timestamp()`, extensions `uuid-ossp`/`pgcrypto`/`pg_trgm`.
  - `.down.sql` drop semua tabel+function+extension.
- Referensi skema asli: `/home/nurdiansyah/Desktop/k-forum-api/internal/migrations/0001_create_auth_tables.up.sql`, `0022_add_account_lockout.up.sql`.

### Fase 4 — Seeder Tooling (global scope only)
- `cmd/seeder/main.go` (port verbatim), `internal/seeders/seeder.go` (registry: `RoleSeeder, PermissionSeeder, RolePermissionSeeder, UserSeeder` saja — drop Subscription/News*/Accounting seeder).
- `internal/seeders/role_seeder.go` — 4 role saja, semua `ScopeType: global`: `usergod` (not assignable), `superadmin`, `admin` (ubah dari region→global), `member`. Drop guest/leader/moderator/community_member.
- `internal/seeders/user_seeder.go` — user bootstrap `usergod@sipon.dev`.
- `internal/seeders/permission_seeder.go` + `internal/seeders/sources/seeder_permissions.json` (rename dari `seeder_permissions_benefits.json`) — hanya 3 permission: `manage_system_settings`, `assign_role`, `manage_users`.
- `internal/seeders/role_permission_seeder.go` — mapping `superadmin`+`usergod` → ketiga permission tsb.

### Fase 5 — Swagger Tooling
- Anotasi di `cmd/app/main.go`: `@title Sipon API`, `@description API development docs untuk Sipon.`, `@BasePath /`, `@securityDefinitions.apikey BearerAuth`.
- Router: swagger UI hanya di-mount saat `appEnv == "development"`.
- `docs/` adalah output generated (`make swagger`), tidak ditulis manual.

### Fase 6 — Makefile
- Target dipertahankan: `dev-up/down`, `dev-all-up/down`, `run`, `migrate-{up,down,fresh,version,force,create}`, `seed{-all,,-role}`, `build`, `tidy`, `test{,-unit,-integration,-usecase}`, `lint`, `swagger{,-check}`.
- Drop: `worker`, `minio-init`, semua `opensearch-*`, semua `observ-*`.
- `migrate-create` **wajib** dipakai untuk migrasi baru setelah genesis migration di Fase 3 (sesuai CLAUDE.md §12).

### Fase 7 — Domain Layer (User, Role/Permission, Verification)
- `internal/domain/user/`: `constant/user_constant.go` (CredentialType hanya LOCAL, LoginIdentifierKind hanya Email/Phone/Username), `entity/{user.go,credential.go,login_identity.go}` (drop UnlinkGoogleCredential, NewGoogleCredential/NewAppleCredential, field IsVerified/AppleRefreshToken), `valueobject/user_vo.go` (port verbatim), `repository/interfaces.go` (drop `SetVerified`).
- `internal/domain/role/`: `constant/role_constant.go` (trim ke 4 role, AdminRoleName scope→global, tetap simpan 3 nilai ScopeType), `entity/{role,permission,role_permission,user_role}.go` (port verbatim — sudah scope-agnostic), `repository/interfaces.go`, `service/user_role_assignment.go` (port verbatim).
- `internal/domain/verification/`: `constant/verification_constant.go` (drop PurposeAccountDeletion), `entity/verification_code.go`, `repository/interfaces.go` — port verbatim.
- `internal/domain/errors/error.go` — port verbatim.
- Referensi: `/home/nurdiansyah/Desktop/k-forum-api/internal/domain/user/`, `.../role/`, `.../verification/`.

### Fase 8 — Infrastructure: Persistence + External Adapters
- `internal/infrastructure/persistence/`:
  - `postgres_transactor.go` — port verbatim.
  - `postgres_user_repository.go` — port minus kolom `is_verified`/`apple_refresh_token`, minus method `SetVerified`. Error-mapping ikut CLAUDE.md §7.
  - `postgres_role_repository.go` — port verbatim (Role/Permission/UserRole/RolePermission CRUD sederhana).
  - `postgres_role_permission_query.go` — port list/pagination methods, **drop** method Community* (community scope out of scope) + tipe port terkait.
  - `postgres_verification_repository.go` (baru, ikuti konvensi `postgres_<context>_repository.go`) — Save/FindLatestByUserAndPurpose/Update untuk `verification_codes`.
- `internal/infrastructure/external/`:
  - `bcrypt/hasher.go` — port verbatim.
  - `jwt/token_generator.go` — port minus seamless token methods.
  - `otpgen/otp_generator.go` — port verbatim.
  - `smtp/{email_sender.go,noop_email_sender.go}` — port minus `SendRegionInvitation`/`SendNotification`.
  - `fonnte/sms_sender.go` — port minus `SendNotification`.
- `internal/infrastructure/cache/`: `redis_session_revocation.go`, `redis_principal_cache.go`, `redis_rate_limiter.go` — port verbatim (semua generic, dipakai penuh oleh auth/RBAC).
- Referensi: `/home/nurdiansyah/Desktop/k-forum-api/internal/infrastructure/persistence/postgres_user_repository.go`, `postgres_role_repository.go`, `internal/infrastructure/external/jwt/token_generator.go`.

### Fase 9 — Application Layer (Ports, Principal, Usecases)
- `internal/app/port/`: `token_generator.go` (minus seamless), `password_hasher.go`, `session_revocation.go`, `principal_cache.go`, `transactor.go`, `otp_generator.go`, `email_sender.go` (trim ke SendOTP+SendPasswordResetOTP), `sms_sender.go` (trim ke SendOTP), `rate_limiter.go`, `role_permission_query_model.go` (drop tipe Community*).
- `internal/app/service/principal/`:
  - `principal.go` — **disederhanakan**: `Principal{UserID, SessionID, Roles[], Permissions[]}` + `HasRole/HasPermission/IsUsergod/IsSuperAdmin/IsAdmin`. Drop `Plan/Benefit/AdminScope` types dan `HasBenefit/ManagedRegionIDs`.
  - `builder.go` — **disederhanakan**: hanya depend ke `UserRepository/UserRoleRepository/RoleRepository/RolePermissionRepository`. Drop subscription/plan/benefit block dan admin_scopes block serta `ScopeNameResolver`.
  - `policy.go` — sisakan `CanActAsOwner`/`CanActAsAuthor` saja (generic, tidak depend ke concept yang di-drop). Drop `CanManageRegion`/`CanModerateInCommunity`/`CanApproveContent`.
- `internal/app/usecase/auth/` (satu file per usecase, sesuai CLAUDE.md §2):
  - `helpers.go` — `issueTokenPair` minus AppleID; drop `generateUniqueUsername`/`makeUsernameBase`/`deriveFullName`/`appleIDOf` (khusus Google/Apple registration).
  - `register.go` — drop subscription/profile/outbox block, tambahkan assign role `member` (global) langsung dalam transaksi.
  - `login.go` — drop outbox publish.
  - `refresh_token.go`, `logout.go` — port verbatim.
  - `session.go` — disederhanakan drastis: hanya `User`, `Roles`, `Permissions` (drop Subscription/Benefits/AdminScopes/Avatar).
  - `me.go` — drop profileRepo/fileUploader/AppleID.
  - `forgot_password.go`, `reset_password.go`, `change_password_local.go`, `set_password_local.go` — port verbatim.
  - `request_identity_otp.go`, `request_change_identity.go`, `confirm_change_identity.go`, `identity_verification.go` — port verbatim.
  - `verify_identity_otp.go` — drop deviceRepo/pushSender/sendEmailVerifiedPush (notifikasi out of scope).
  - **Tidak diporting**: `google_login.go`, `apple_login.go`, `generate_seamless_token.go`, `verify_seamless_token.go`.
- `internal/app/usecase/rolepermission/` — port **seluruhnya verbatim** (sudah scope-agnostic, tidak perlu perubahan logic): `dependencies.go`, `helpers.go`, dan seluruh file create/update/delete/get/list untuk permission/role/role_permission/user_role.
- Referensi: `/home/nurdiansyah/Desktop/k-forum-api/internal/app/service/principal/builder.go`, `internal/app/usecase/auth/register.go`, `internal/app/usecase/rolepermission/dependencies.go`.

### Fase 10 — HTTP Layer (middleware, handler, router)
- `internal/interfaces/http/middleware/`: `auth.go` (JWTAuth, OptionalJWTAuth, PrincipalLoader, GetPrincipal, RequireRole, RequirePermission, RequireAdmin, RequireSuperAdmin — port verbatim, **drop RequireBenefit**), `cors.go`, `request_id.go`, `request_logger.go`, `rate_limit.go` — port verbatim. Drop `locale.go` dan `maintenance.go`.
- `internal/interfaces/http/httperror/`: `middleware.go` port verbatim; `http_error.go` — drop i18n bundle, `Handle` pakai message langsung; `validation.go` — cek dependency ke i18n bundle, lepas jika ada.
- `internal/interfaces/http/respond/` — port verbatim (`Success`, `SuccessWithMeta`, `Error`, dll — sesuai CLAUDE.md §8).
- `internal/app/dto/`: `pagination_dto.go` (verbatim), `auth_dto.go` (drop Google/Apple/Seamless/ListUsers DTO, drop field Avatar/AppleID/PendingAcceptances), `session_dto.go` (disederhanakan sesuai Principal baru), `role_permission_dto.go` (port verbatim penuh).
- `internal/interfaces/http/handler/web/`: `auth_handler.go` (drop GoogleLogin/ListUsers/VerifySeamlessToken handler+swagger annotation terkait), `role_permission_handler.go` (port verbatim penuh).
- `internal/interfaces/http/router/router.go` — hanya group web (drop mobileAuth/fcm/user-settings), sesi (`GET /auth/session`, `POST /auth/logout`), protected web auth (`me`, change-password, set-password, change-email/phone request+confirm), group `/role-permission` dengan `RequireRole("superadmin","usergod")` (17 route seperti source). Drop `MaintenanceGate`, `Locale()`.
- Referensi: `/home/nurdiansyah/Desktop/k-forum-api/internal/interfaces/http/router/router.go`, `internal/interfaces/http/middleware/auth.go`, `internal/interfaces/http/handler/web/{auth_handler.go,role_permission_handler.go}`.

### Fase 11 — Tests
- `internal/testhelper/`: `testdb.go` (testcontainers postgres, run migration via `migrations.FS`), `noop.go` (`noopPrincipalCache`, `noopSMSSender`, `fakeSessionRevocationStore`+constructor), `seed.go` (`MustSeedNamedRole`, `MustSeedPermissionForRole`, `MustVerifyUserEmailIdentity`, `MustAssignUserRole` — drop MustSeedDefaultPlan/PlanBenefit/AccountingSettings), `testserver.go` (`MustStartTestServer` — wire semua repo/usecase/handler/router versi simplified di atas), `fixtures.go` (cek isi relevan saat implementasi).
- `internal/interfaces/http/handler/web/`: `main_test.go`, `testutil_test.go` (port helper `mustRegisterUser`/`mustLogin`/`mustGetOTPFromDB`/dll — baca source file penuh saat implementasi untuk signature persis), `auth_handler_test.go` (port verbatim — semua case sudah dalam scope), `role_permission_handler_test.go` (**baru** — cover 5 skenario wajib CLAUDE.md §11: 401 unauthenticated, 422 invalid payload, 404 not found, 403 forbidden non-superadmin, 2xx success CRUD penuh).
- Referensi: `/home/nurdiansyah/Desktop/k-forum-api/internal/testhelper/testserver.go`, `internal/interfaces/http/handler/web/auth_handler_test.go`.

---

## Ringkasan Rename "K-Forum → Sipon"

| Lokasi | Lama | Baru |
|---|---|---|
| `go.mod` module | `k-forum-api` | `sipon-api` |
| Semua Go import | `k-forum-api/internal/...` | `sipon-api/internal/...` |
| Docker network | `k-forum-net` | `sipon-net` |
| Container names | `k-forum-api_postgres` dst | `sipon-api_postgres` dst |
| `.env.example` DB default | `DB_USER=k-forum`, `DB_NAME=k-forum` | `DB_USER=sipon`, `DB_NAME=sipon` |
| `cmd/app/main.go` swagger | `@title K-Forum API` | `@title Sipon API` |
| `internal/seeders/user_seeder.go` | `usergod@k-forum.dev` | `usergod@sipon.dev` |
| `internal/seeders/seeder.go` query | `WHERE email = 'usergod@k-forum.dev'` | `WHERE email = 'usergod@sipon.dev'` |
| SMTP `From` default | — | `noreply@sipon.dev` |
| `Makefile` help banner | `k-forum-api — perintah yang tersedia:` | `sipon-api — perintah yang tersedia:` |

Tidak ditemukan string "K-Forum"/"k-forum" lain di domain/usecase/infra/handler code — branding hanya ada di config default, docker naming, seeder, dan swagger annotation di atas.

---

## Verifikasi per Fase

- Setiap fase: `go build ./...` harus sukses (begitu ada file Go).
- Setelah Fase 3: `make migrate-up` (via docker compose) berhasil membuat seluruh tabel di Postgres kosong.
- Setelah Fase 4: `make seed-all` idempotent — bisa dijalankan berulang tanpa error, dan `usergod@sipon.dev` + role `usergod` ter-assign.
- Setelah Fase 5: `make swagger` menghasilkan `docs/swagger.json` valid dengan title "Sipon API"; `GET /swagger/*any` bisa diakses saat `APP_ENV=development`.
- Setelah Fase 10: `make run` (atau `make dev-all-up` lalu curl) — `POST /api/v1/web/auth/register` dan `/login` berhasil mengembalikan token; `GET /api/v1/auth/session` mengembalikan roles+permissions yang benar untuk user yang baru register (role `member`).
- Setelah Fase 11: `go test ./internal/interfaces/http/handler/web/... -count=1 -timeout 300s` hijau semua, termasuk `role_permission_handler_test.go` yang baru dibuat; `go test ./internal/domain/... -short` dan `go test ./internal/infrastructure/persistence/... -timeout 120s` (butuh Docker testcontainers) juga hijau.
