# Plan: Modular Monolith DDD — `auth` & `kesantrian`

> Restructure `sipon-api` dari layered DDD (`internal/{domain,app,infrastructure,interfaces}`) menjadi **modular monolith** dengan dua module vertikal: `auth` dan `kesantrian`. Domain `santri` dan semua yang mengikutinya masuk module `kesantrian`; sisanya (`user`, `role`, `verification`) masuk module `auth`. Shared kernel di `internal/shared`.

---

## Keputusan final

1. **Struktur folder** — `internal/modules/{auth,kesantrian}/{domain,application,infrastructure,interfaces}` + `internal/shared`. Tiap module memajangkan semua layer-nya sendiri (vertical slice).
2. **Coupling kesantrian → auth** — gunakan **anti-corruption port** `UserAccountPort` (+ `UserSummaryReadModel`) yang **didefinisikan di module kesantrian** dan diimplementasi di module auth (consumer defines port; satu arah: auth import tipe port dari kesantrian, kesantrian tidak tahu auth).
3. **NIS** — `kesantrian` pemilik otoritatif format/validasi NIS. Auth hanya menerima string opaque lewat `LoginIdentifierNIS`. Hapus regex `nisPatternRe` & deteksi NIS-pattern di `user/valueobject.LoginIdentifier` (auto-derive gender dsb. tidak ada di auth).
4. **Migration** — `cmd/migrate` beralih ke source **iofs** yang meng-embed kedua folder migration module (`modules/auth/.../migrations` + `modules/kesantrian/.../migrations`). Migration `20260729065536_add_santri_nis_and_requests` dipecah:
   - Bagian `ALTER user_identities` (tambah kind `'NIS'` ke CHECK) → migration **auth**.
   - Bagian `ALTER santri ADD nis` + `santri_requests` → migration **kesantrian**.
5. **Langkah refactor** — **full refactor sekali jalan**; build + vet + test harus hijau di akhir. Tidak ada perubahan behavior.

---

## A. Struktur Target

```
internal/
  shared/                          ← shared kernel (di-import kedua module)
    domainerror/                   (ex internal/domain/errors; package `domainerror`)
    apperror/                      (ex internal/app/apperror)
    app/
      dto/pagination.go            (ex internal/app/dto/pagination_dto.go)
      port/                        (10 port lintas module: transactor, password_hasher,
                                     token_generator, session_revocation, principal_cache,
                                     rate_limiter, otp_generator, email_sender, sms_sender,
                                     file_uploader)
      principal/                   (TYPE saja: Principal/Role/Permission/UserScope —
                                     dipisah dari Builder agar tidak import module)
      service/media/               (object path helpers generic)
    config/
    logger/
    infrastructure/
      cache/                       (Redis principal cache, rate limiter, session revocation)
      external/                    (bcrypt, jwt, fonnte, smtp, otpgen, minio)
      persistence/
        postgres_transactor.go     (+ exported TxFromContext / NewTxContext)
        postgres_pagination.go
    interfaces/http/
      httperror/  respond/  middleware/{cors,rate_limit,request_id,request_logger}
    migrations/  (embed.go — iofs gabungan auth+kesantrian)
    testhelper/  (testdb, assert, noop, txhelper, seed)

  modules/
    auth/
      domain/  {user, role, verification}   (ex internal/domain/{user,role,verification})
      application/
        dto/                       (auth, profile, session, role_permission, user_management)
        usecase/  {auth, rolepermission, usermanagement}
        service/principal/         (Builder — auth-owned, butuh user+role repos)
        port/                      (user_query_model.go, role_permission_query_model.go)
      infrastructure/
        persistence/               (postgres_{user,verification,role,user_role,
                                     role_permission,role_scope,role_query,user_query})
        migrations/                (auth-owned SQL; termasuk alter NIS kind)
      interfaces/http/
        handler/web/               (auth, role_permission, user_management handlers + tests)
        middleware/auth.go         (JWTAuth + PrincipalLoader — butuh principal.Builder)
      seeders/                     (role_seeder, user_seeder)
      testutil/                    (wiring repos/auth + principal builder + cache noop)

    kesantrian/
      domain/santri/               (ex internal/domain/santri — NIS VO otoritatif di sini)
      application/
        dto/                       (dto santri yang saat ini di usecase/santri)
        usecase/santri/
        port/                      (user_account.go, user_summary_read_model.go —
                                     definisi port yang diimplementasi auth)
      infrastructure/
        persistence/               (postgres_santri_{repository,dokumen,request})
        migrations/                (santri_* tables only)
      interfaces/http/
        handler/web/santri_handler.go
      testutil/                    (wiring repos santri + stub UserAccountPort)

cmd/
  app/main.go                     (composition root; wiring manual tetap, struktur dipecah)
  migrate/main.go                 (switch ke source iofs gabungan)
  seeder/main.go                  (memanggil auth.Seeders; tidak ada santri seeder)
```

---

## B. Shared Kernel (`internal/shared`)

Dipindahkan dari lokasi lama:

| Lama | Baru | Catatan |
|---|---|---|
| `internal/domain/errors` | `internal/shared/domainerror` | package `domainerror`; update semua alias `domainerr` |
| `internal/app/apperror` | `internal/shared/apperror` | — |
| `internal/app/dto/pagination_dto.go` | `internal/shared/app/dto/pagination.go` | — |
| `internal/app/port/{transactor,password_hasher,token_generator,session_revocation,principal_cache,rate_limiter,otp_generator,email_sender,sms_sender,file_uploader}.go` | `internal/shared/app/port/` | 10 port lintas module |
| `internal/app/service/principal` (type `Principal/Role/Permission/UserScope`) | `internal/shared/app/principal` | **type saja** — supaya tidak cyclic dengan module |
| `internal/app/service/principal/builder.go` | `internal/modules/auth/application/service/principal/builder.go` | Builder butuh user+role repos → auth-owned |
| `internal/app/service/media` | `internal/shared/app/service/media` | — |
| `internal/config` | `internal/shared/config` | — |
| `internal/logger` | `internal/shared/logger` | — |
| `internal/infrastructure/cache` | `internal/shared/infrastructure/cache` | — |
| `internal/infrastructure/external/{bcrypt,jwt,fonnte,smtp,otpgen,minio}` | `internal/shared/infrastructure/external/...` | — |
| `internal/infrastructure/persistence/postgres_transactor.go`, `postgres_pagination.go` | `internal/shared/infrastructure/persistence/...` | + `TxFromContext`, `NewTxContext` |
| `internal/interfaces/http/httperror` | `internal/shared/interfaces/http/httperror` | — |
| `internal/interfaces/http/respond` | `internal/shared/interfaces/http/respond` | — |
| `internal/interfaces/http/middleware/{cors,rate_limit,request_id,request_logger}` | `internal/shared/interfaces/http/middleware/...` | — |
| `internal/interfaces/http/middleware/auth.go` | `internal/modules/auth/interfaces/http/middleware/auth.go` | butuh `principal.Builder` (auth-owned) |
| `internal/migrations/embed.go` | `internal/shared/migrations/embed.go` | iofs embed **dua** folder module |
| `internal/testhelper/{testdb,assert,noop,txhelper,seed}.go` | `internal/shared/testhelper/...` | — |

---

## C. Module `kesantrian`

Sumber: `internal/domain/santri/*`, `internal/app/usecase/santri/*`, `internal/infrastructure/persistence/postgres_santri_*`, `internal/interfaces/http/handler/web/santri_handler.go`.

### C.1 Port anti-corruption (`modules/kesantrian/application/port/`)

**`UserAccountPort`** — menggantikan langsung import `domain/user` di usecase santri:

```go
package port

type CreateUserForSantriRequest struct {
    NIS       string
    Gender    string // derived from NIS di kesantrian
    FullName  string
    // field lain sesuai kebutuhan create_santri
}
type CreateUserForSantriResult struct {
    UserID          string
    GeneratedPassword string
}
type UserAccountPort interface {
    CreateUserForSantri(ctx context.Context, req CreateUserForSantriRequest) (CreateUserForSantriResult, error)
    AttachNISLoginIdentity(ctx context.Context, userID string, nis string) error
}
```

- `CreateUserForSantri` menggantikan body `create_santri.go:30-108` (membangun `User`+`Credential`+3 `LoginIdentity` NIS/EMAIL/USERNAME + `userRepo.Save`).
- `AttachNISLoginIdentity` menggantikan body `approve_santri_request.go:74-87` (load user → add NIS `LoginIdentity` → `userRepo.Update`).

**`UserSummaryReadModel`** — menggantikan `userRepo.FindByID` di enrichment read:

```go
type UserSummary struct {
    UserID      string
    Username    string
    Email       string
    PhoneNumber string
    AvatarKey   string
}
type UserSummaryReadModel interface {
    FindSummaries(ctx context.Context, userIDs []string) (map[string]UserSummary, error)
}
```

Dipakai oleh `get_santri`, `update_santri`, `list_santri`, `list_santri_requests`.

### C.2 Perubahan `Dependencies` usecase santri

Lama:
```go
type Dependencies struct {
    SantriRepo      santrirepo.SantriRepository
    SantriDokumenRepo ...
    SantriRequestRepo ...
    UserRepo        userrepo.UserRepository   // ← dihapus
    FileUploader    port.FileUploader
    Hasher          port.PasswordHasher        // ← dihapus (hashing pindah ke auth impl)
    Transactor      port.Transactor            // ← pertahankan kalau masih ada tx santri+santri_dokumen
}
```
Baru:
```go
type Dependencies struct {
    SantriRepo      santrirepo.SantriRepository
    SantriDokumenRepo ...
    SantriRequestRepo ...
    UserAccount     port.UserAccountPort       // ← baru
    UserSummaries   port.UserSummaryReadModel  // ← baru
    FileUploader    sharedport.FileUploader
    Transactor      sharedport.Transactor      // jika masih diperlukan (periksa usecase)
}
```

### C.3 NIS otoritatif

`modules/kesantrian/domain/santri/valueobject.NIS` tetap satu-satunya tempat validasi format NIS (`^1000[12][0-9]{5}$`) & derive gender. Kesantrian meneruskan string NIS ke port; auth tidak memvalidasi.

---

## D. Module `auth`

Sumber: `internal/domain/{user,role,verification}/*`, `internal/app/usecase/{auth,rolepermission,usermanagement}/*`, `app/service/principal/builder.go`, semua `postgres_{user,verification,role,role_permission,role_scope,role_query,user_query}`, 3 handler auth-side.

### D.1 Implementasi port kesantrian

Di `modules/auth/application/service/` (atau `application/portadapter/`):

- `user_account_service.go` — implementasi `kesantrian/application/port.UserAccountPort`:
  - `CreateUserForSantri` — pindahkan body `create_santri.go:30-108`. Boleh import `domain/user` + `shared/app/port.{PasswordHasher,Transactor}`. Untuk atomicity tx, baca `TxFromContext(ctx)` (tx terbuka lewat context dari kesantrian).
  - `AttachNISLoginIdentity` — pindahkan body `approve_santri_request.go:74-87`.
- `user_summary_read_model.go` — implementasi `kesantrian/application/port.UserSummaryReadModel` — wrapper `userRepo.FindByID` atau batch query baru.

Catatan dependency direction: auth import **tipe port** dari `modules/kesantrian/application/port` (inverted dependency). Tidak ada coupling balik: kesantrian tidak import auth.

### D.2 NIS di auth menjadi opaque

- `modules/auth/domain/user/valueobject/user_vo.go`:
  - Hapus `nisPatternRe` regex.
  - `LoginIdentifier` tidak lagi auto-detect NIS-pattern.
  - `LoginIdentifierNIS` kind tetap ada (dipasang oleh impl `CreateUserForSantri`), tapi nilai disimpan tanpa validasi format & tanpa auto-derive gender.
- `domain/user/constant`: `LoginIdentifierNIS` tetap.

---

## E. Migration & Schema

| File lama | Module pemilik baru | Catatan |
|---|---|---|
| `20260726120000_create_auth_rbac_tables` | auth | `users`, `credentials`, `user_identities`, `roles`, `user_roles`, `verification_codes` |
| `20260726150000_create_role_permissions_table` | auth | `role_permissions` |
| `20260728125618_create_role_scopes` | auth | `role_scopes` |
| `20260728052250_add_avatar_key_to_users` | auth | `ALTER users` |
| `20260729032931_create_santri_tables` | kesantrian | `santri`, `santri_dokumen` |
| `20260729065536_add_santri_nis_and_requests` | **dipecah** | (1) `ALTER user_identities` add NIS kind → **auth**; (2) `ALTER santri ADD nis` + `santri_requests` → **kesantrian** |

### E.1 iofs embed gabungan

`internal/shared/migrations/embed.go`:
```go
package migrations

import (
    "embed"
    "io/fs"

    //go:embed all/modules/auth/infrastructure/migrations
    //go:embed all/modules/kesantrian/infrastructure/migrations
)
var embedFS embed.FS

func Sources() (fs.FS, error) {
    return fs.Sub(embedFS, ".")
}
```
`cmd/migrate/main.go` beralih dari `migrate.New("file://"+dir,...)` ke `migrate.NewFromSource(migrations.Sources(),...)` (iofs source). Tes urutan timestamp agar tidak ada tabrakan.

### E.2 Kontrak schema lintas-module (single-DB modular monolith)

- `santri.user_id REFERENCES users(id) ON DELETE CASCADE` — FK ke tabel auth-owned; wajar di modular monolith single-DB.
- `santri_requests.user_id` & `santri_requests.reviewed_by` → `users(id)`.
- `user_identities` tabel auth-owned; kesantrian hanya **menulis** baris NIS lewat port auth (tidak menyentuh SQL langsung).

---

## F. Test split

Pecah `internal/testhelper/testserver.go` (saat ini wiring semua module):

- `internal/shared/testhelper/` — `testdb.go` (testcontainers), `assert.go` (`AssertDomainError` pakai `shared/domainerror`), `noop.go` (noop principal cache, sms sender, fake session revocation), `txhelper.go` (persistence tx helpers), `seed.go` (raw SQL role seed).
- `internal/modules/auth/testutil/` — wiring repos auth + principal builder + cache noop + use cases auth.
- `internal/modules/kesantrian/testutil/` — wiring repos santri + **stub** `UserAccountPort`/`UserSummaryReadModel`.
- `internal/shared/testhelper/testserver.go` (composer) — bangun server test lintas module (gabung modul auth+kesantrian), router composer.
- Test di `handler/web/*_test.go` pindah bersama handler masing-masing module.

---

## G. Composition root (`cmd/app/main.go`)

Wiring manual DI tetap, urutan dipecah:

1. **Shared infra** — `config.Load()`, `logger.New`, `sql.Open("pgx",...)`, `redis.NewClient`, external services (`bcrypt`, `extjwt`, `smtp`, `fonnte`, `otpgen`, `minio`).
2. **Auth repos** — userRepo, verifRepo, roleRepo, userRoleRepo, rolePermissionRepo, roleScopeRepo, rolePermissionReadModel, userReadModel.
3. **Auth external/use cases** — hashing, jwt; 21 `NewXUseCase` auth; `principal.NewBuilder(userRepo, userRoleRepo, roleRepo, rolePermissionRepo, roleScopeRepo)`; `cache.NewRedisPrincipalCache`, `cache.NewRedisRateLimiter`, `cache.NewRedisSessionRevocationStore`.
4. **Auth port adapter** — `UserAccountService` (impl `UserAccountPort`) + `UserSummaryReadModelImpl` (impl `UserSummaryReadModel`). Inject `userRepo`, `hasher`, `transactor`.
5. **Kesantrian repos** — santriRepo, santriDokumenRepo, santriRequestRepo.
6. **Kesantrian use cases** — `santri.NewUseCases(Dependencies{...})` dengan `UserAccount: UserAccountService`, `UserSummaries: UserSummaryReadModelImpl`.
7. **Rolepermission + usermanagement use cases** (auth).
8. **Handlers** — `NewAuthHandler`, `NewRolePermissionHandler`, `NewUserManagementHandler`, `NewSantriHandler`.
9. **Router composer** — `auth.RegisterRoutes(rg, deps)` + `kesantrian.RegisterRoutes(rg, deps)` + shared middleware (JWT/PrincipalLoader di-attach sebelum route protected). Ganti satu `router.Setup` monolitik.
10. **HTTP server** + graceful shutdown.

`cmd/seeder/main.go` — panggil `auth.Seeders`; tidak ada santri seeder.

---

## H. Import path rewriting

Setelah pemindahan fisik, jalankan `goimports -w .` (atau `gofmt -w` + perbaiki import manual). Mapping utama:

| Lama | Baru |
|---|---|
| `sipon-api/internal/domain/errors` | `sipon-api/internal/shared/domainerror` |
| `sipon-api/internal/app/apperror` | `sipon-api/internal/shared/apperror` |
| `sipon-api/internal/app/dto/pagination_dto` | `sipon-api/internal/shared/app/dto` |
| `sipon-api/internal/app/port/{10 shared}` | `sipon-api/internal/shared/app/port` |
| `sipon-api/internal/app/service/principal` (type) | `sipon-api/internal/shared/app/principal` |
| `sipon-api/internal/app/service/principal` (Builder) | `sipon-api/internal/modules/auth/application/service/principal` |
| `sipon-api/internal/app/service/media` | `sipon-api/internal/shared/app/service/media` |
| `sipon-api/internal/config` | `sipon-api/internal/shared/config` |
| `sipon-api/internal/logger` | `sipon-api/internal/shared/logger` |
| `sipon-api/internal/infrastructure/{cache,external,persistence}` | `sipon-api/internal/shared/infrastructure/...` |
| `sipon-api/internal/interfaces/http/{httperror,respond,middleware/{cors,rate_limit,request_id,request_logger}}` | `sipon-api/internal/shared/interfaces/http/...` |
| `sipon-api/internal/interfaces/http/middleware/auth` | `sipon-api/internal/modules/auth/interfaces/http/middleware` |
| `sipon-api/internal/domain/{user,role,verification}` | `sipon-api/internal/modules/auth/domain/{user,role,verification}` |
| `sipon-api/internal/app/usecase/{auth,rolepermission,usermanagement}` | `sipon-api/internal/modules/auth/application/usecase/{auth,rolepermission,usermanagement}` |
| `sipon-api/internal/app/dto/{auth,profile,session,role_permission,user_management}` | `sipon-api/internal/modules/auth/application/dto` |
| `sipon-api/internal/domain/santri` | `sipon-api/internal/modules/kesantrian/domain/santri` |
| `sipon-api/internal/app/usecase/santri` | `sipon-api/internal/modules/kesantrian/application/usecase/santri` |
| `sipon-api/internal/infrastructure/persistence/postgres_user_*` | `sipon-api/internal/modules/auth/infrastructure/persistence` |
| `sipon-api/internal/infrastructure/persistence/postgres_santri_*` | `sipon-api/internal/modules/kesantrian/infrastructure/persistence` |
| `sipon-api/internal/interfaces/http/handler/web/{auth,role_permission,user_management}_handler` | `sipon-api/internal/modules/auth/interfaces/http/handler/web` |
| `sipon-api/internal/interfaces/http/handler/web/santri_handler` | `sipon-api/internal/modules/kesantrian/interfaces/http/handler/web` |
| `sipon-api/internal/interfaces/http/router` | dihapus — diganti composer per module |
| `sipon-api/internal/migrations` | `sipon-api/internal/shared/migrations` (embed) |
| `sipon-api/internal/seeders` | `sipon-api/internal/modules/auth/seeders` |
| `sipon-api/internal/testhelper` | `sipon-api/internal/shared/testhelper` + per-module `testutil` |

Catatan package name: folder `domainerror` → `package domainerror`. Update alias `domainerr` di seluruh repo (sudah dipakai di `santri/constant` dsb.) konsisten ke `domainerror`.

---

## I. Risiko & Mitigasi

- **Tx cross-module** — `UserAccountPort` menerima `ctx` ber-transaksi; impl di auth membaca `TxFromContext(ctx)` sehingga atomicity `santri_repo.Save` + `user_repo.Save` tetap dalam satu DB transaction. Pastikan port meneruskan `ctx` ber-transaksi (jangan `context.Background()` di impl).
- **Import cycle** — `shared/app/principal` hanya tipe (tidak import module); `Builder` di auth import shared type. `UserAccountPort` di kesantrian; auth import tipe port-nya. Kesantrian tidak import auth → tidak cycle.
- **Embed path** — `//go:embed` path relatif terhadap `internal/shared/migrations/embed.go`. Path `all/modules/auth/infrastructure/migrations` tidak akan resolve (embed path relatif terhadap file `.go`, tidak bisa naik ke atas `internal`). Solusi: pindahkan `embed.go` ke root module (`internal/modules/<m>/infrastructure/migrations/embed.go`) atau gunakan `//go:embed modules/...` dari repo root dengan `embedFS` di package `cmd/app`/`internal/migrations`. **Final: package `migrations` ditaruh di repo-root-level** (`internal/migrations/embed.go`) agar dapat `//go:embed modules/auth/infrastructure/migrations/*.sql modules/kesantrian/infrastructure/migrations/*.sql`; `internal/shared/migrations` hanya menyimpan helper bila perlu. Periksa saat implementasi & sesuaikan path embed.
- **Timestamp migration tabrakan** — saat pecah `20260729065536`, jangan reuse timestamp sama. Buat timestamp baru (mis. `20260729065536_auth_add_nis_identity_kind` + `20260729065537_santri_add_nis_and_requests`) supaya unik & urutan aman.
- **`Hasher` drop di kesantrian** — verifikasi tidak ada usecase santri lain yang pakai `Hasher` selain `create_santri`. Begitu juga `Transactor` — periksa pakai di usecase santri mana saja (`create_santri` tx user+santri → jadi tanggung jawab impl port di auth; tx santri_repo+santri_dokumen/requests mungkin masih perlu di kesantrian).
- **`principal.Builder` auth-owned** — `middleware/auth.go` pindah ke module auth; route composer di `cmd/app` pasang middleware ini. Port `PrincipalCachePort` di shared menerima tipe `*principal.Principal` dari shared (bukan Builder).

---

## J. Urutan Eksekusi

1. Buat struktur folder baru `internal/shared/*` & `internal/modules/{auth,kesantrian}/*`.
2. Pindahkan & rename shared kernel (§B). Perbaiki import.
3. Pindahkan module auth (domain, application, infrastructure, interfaces, migration, seeders). Perbaiki import.
4. Definisikan `UserAccountPort` + `UserSummaryReadModel` di `modules/kesantrian/application/port/`.
5. Implementasi di module auth: `UserAccountService` + `UserSummaryReadModel` impl (menerima `ctx` ber-tx).
6. Hapus regex NIS & deteksi pattern di `user/valueobject/user_vo.go`; pertahankan `LoginIdentifierNIS` kind.
7. Pindahkan module kesantrian (domain + NIS VO otoritatif, application usecase, infrastructure, interfaces). Hapus semua import `domain/user` dari usecase santri; ganti dengan `UserAccountPort`/`UserSummaryReadModel`. Update `Dependencies`.
8. Migration parallel split (auth `add_nis_identity_kind` + kesantrian `santri_add_nis_and_requests`); setup iofs embed + switch `cmd/migrate` ke source iofs.
9. Pecah `testhelper/testserver.go` → shared testhelper + per-module testutil + composer.
10. Restrukturisasi `cmd/app/main.go` (komposisi module + per-module `RegisterRoutes`) + `cmd/seeder` panggil `auth.Seeders`.
11. `goimports -w .` → `go build ./...` → `go vet ./...` → `go test ./...`. Iterasi sampai hijau. Tidak ada perubahan behavior.

---

## K. Verifikasi

```bash
goimports -w .
go build ./...
go vet ./...
go test ./...
```

Target: build hijau, semua test existing lulus, tidak ada perubahan behavior fungsional. Satu-satunya perubahan behavioracceptable: hapus validasi format NIS di auth (sebelumnya duplikat) — tidak mengubah alur fungsional karena kesantrian tetap memvalidasi.