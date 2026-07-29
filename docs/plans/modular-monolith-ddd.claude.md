# Restructure sipon-api → Modular Monolith (modules: `auth`, `kesantrian`)

## Context

sipon-api saat ini adalah DDD **layer-first**: satu `internal/domain/*` per bounded context (santri, user, role, verification), tapi `internal/app/usecase/*` per modul-usecase, dan `internal/infrastructure/persistence` + `internal/interfaces/http/handler/web` + `internal/interfaces/http/router/router.go` semuanya **flat** — satu package berisi semua domain sekaligus, dibedakan hanya lewat prefix nama file. Setelah domain `santri` (kesantrian) ditambahkan (commit `a3b81ac`), campuran ini makin terasa: `router.go` dan `handler/web` menumpuk 4 domain dalam satu file/package, dan santri usecase sudah menembus langsung ke `domain/user` untuk membuat/mengubah akun.

User ingin restructure ke **modular monolith**: domain `santri` dan semua yang mengikutinya (usecase, persistence, handler) masuk ke module `kesantrian`; sisanya (user, role/permission/scope, verification/OTP, usecase auth) masuk ke module `auth`. DDD tetap dipertahankan **di dalam** tiap module (domain/usecase/infrastructure/interfaces empat layer, aturan dependency CLAUDE.md §3 tetap berlaku, hanya sekarang bersarang di bawah module).

Keputusan yang sudah dikonfirmasi user:
1. **Cross-module dependency** (kesantrian butuh referensi ke User milik auth): pakai **port lokal + adapter (anti-corruption layer)** — kesantrian TIDAK import domain/app auth secara langsung.
2. **Struktur folder**: `internal/modules/{auth,kesantrian}/...`.
3. **Test infra**: tetap **satu shared TestServer/harness** (bukan dipisah per module) — masih satu binary/satu DB.

Rencana ini sudah diverifikasi terhadap kode aktual (`postgres_transactor.go`, `create_santri.go`, `router.go` sudah dibaca langsung) — bukan asumsi dari eksplorasi awal saja.

---

## Temuan penting yang membentuk urutan kerja

1. **`internal/infrastructure/persistence` adalah satu package Go** dengan helper package-private yang dipakai LINTAS domain: `execFromContext`/`txKey`/`executor` (di `postgres_transactor.go`), `resolvePaginationParams` (`postgres_pagination.go`), `nullableString` (di `postgres_role_repository.go`). Ini **harus di-export** ke package platform bersama SEBELUM file persistence dipecah per module, kalau tidak build langsung rusak.
2. **`internal/domain/errors` pakai `package errors`** (shadow stdlib) — tapi semua pemanggil sudah alias `domainerr`, jadi rename folder ke `domainerr` = zero identifier churn.
3. **Kesantrian sudah menembus domain auth lebih dalam dari sekadar "baca ID"**: `create_santri.go` membangun `userentity.User`, credential lokal, 3 `LoginIdentity` (NIS/email/username), generate+hash password, lalu `userRepo.Save` — logika provisioning akun ini **milik auth**, bukan kesantrian, dan harus dipindah ke sana di balik port. `approve_santri_request.go` juga menulis ulang credential/identity user. `update_santri.go` mengubah `user.Fullname`. `get_santri.go`/`list_santri.go`/`list_santri_requests.go` membaca field user (N+1 query per baris — sekalian diperbaiki jadi batch).
4. **`middleware.JWTAuth`** butuh `port.TokenGenerator` + `port.SessionRevocationStore` → keduanya harus **shared**, bukan milik module manapun, supaya platform HTTP kernel tidak import module apapun.
5. **`internal/app/service/principal`** dipakai 9 file termasuk yang akan jadi platform (`middleware/auth.go`, `router.go`) — dipecah: tipe (`Principal`, `Role`, dst) → shared `authz` package; `builder.go` (butuh repo user+role) → tetap di module `auth`, diekspos ke platform lewat interface `authz.Resolver`.
6. **Router meng-guard route admin santri pakai `roleconstant.PermissionManageUsers`** (constant milik auth) — didokumentasikan sebagai pengecualian sementara (module kesantrian boleh import `auth/domain/role/constant` untuk permission key saja), dengan catatan perbaikan jangka panjang (permission registry) di CLAUDE.md.
7. **`internal/domain/user/constant/user_constant.go` punya `CodeInvalidNISFormat` yang TIDAK dipakai di manapun** (duplikat dari punya santri) — dihapus saat migrasi, bukan masalah fungsional tapi bersihkan sekalian.
8. **Migrasi tetap flat** di `internal/migrations/` — jangan dipindah/dipecah per module (Dockerfile & `Makefile migrate-create` bergantung ke path ini; CLAUDE.md §12 juga sudah fix soal ini).
9. **`Makefile`** punya target test yang path-nya akan diam-diam kosong (`./internal/domain/...`, `./internal/infrastructure/persistence/...`, `./internal/app/...`) — wajib diupdate di akhir, kalau tidak safety net hilang tanpa disadari.
10. Santri belum punya `santri_handler_test.go` (celah CLAUDE.md §11 yang sudah ada sebelum restructure ini) — jadi kesempatan bagus untuk ditutup sekalian saat handler-nya dipindah ke module baru.

---

## Target struktur direktori

```
internal/
├── migrations/                                   # TIDAK PINDAH (flat, timestamp-based, sudah benar)
│
├── shared/                                        # kontrak & tipe murni, tanpa I/O — leaf, tidak boleh import siapa pun
│   ├── domain/domainerr/error.go                  # package domainerr (dari internal/domain/errors)
│   └── app/
│       ├── apperror/app_error.go                  # package apperror
│       ├── dto/pagination_dto.go                  # package dto (PaginationParams, Meta — HANYA yang generic)
│       ├── media/object_path.go                   # package media (type ObjectPath string saja)
│       └── port/                                  # port yang genuinely lintas-module (technology-facing)
│           ├── email_sender.go, file_uploader.go, otp_generator.go, password_hasher.go
│           ├── rate_limiter.go, session_revocation.go, sms_sender.go
│           ├── token_generator.go, transactor.go
│   └── authz/
│       ├── principal.go                           # package authz (Principal, Role, Permission, UserScope + method)
│       ├── cache.go                               # package authz (Cache interface, dulu port.PrincipalCachePort)
│       └── resolver.go                            # NEW: type Resolver interface{ Build(ctx, userID, sessionID) (*Principal, error) }
│
├── platform/                                      # implementasi teknis + kernel HTTP/DB/CLI — tidak boleh import modules/**
│   ├── config/config.go
│   ├── logger/{logger.go,context.go,context_handler.go}
│   ├── database/
│   │   ├── transactor.go                          # package database: Transactor (dulu PostgresTransactor), TxFromContext, NewTxContext, Executor, ExecFromContext (export dari execFromContext)
│   │   ├── pagination.go                          # ResolvePaginationParams (export dari resolvePaginationParams)
│   │   └── nullable.go                             # NEW: NullableString (export dari nullableString)
│   ├── cache/{redis_principal_cache.go,redis_rate_limiter.go,redis_session_revocation.go}
│   ├── external/{bcrypt,fonnte,jwt,minio,otpgen,smtp}/*.go
│   ├── http/
│   │   ├── httperror/{http_error.go,middleware.go,validation.go}
│   │   ├── middleware/{auth.go,cors.go,rate_limit.go,request_id.go,request_logger.go}   # auth.go pakai authz.Resolver + authz.Cache, bukan principal.Builder
│   │   ├── respond/{error.go,success.go}
│   │   └── router/router.go                       # REWRITE: Kernel{...} + Groups{...} + type RouteRegistrar interface{ RegisterRoutes(Groups) } + func New(Kernel, ...RouteRegistrar) *gin.Engine — TIDAK import modules apapun
│   └── seeder/seeder.go                            # package seeder (dari seeders), RunAll/RunByName jadi variadic(seeders ...Seeder)
│
├── modules/
│   ├── auth/
│   │   ├── module.go                               # package auth: Deps, Module{Handlers, Principal authz.Resolver, UserDirectory api.UserDirectory}, New(Deps) *Module, RegisterRoutes
│   │   ├── api/user_directory.go                   # package api: UserDirectory interface (FindAccount, FindAccounts batch, UpdateFullname, ProvisionAccount, AttachNISIdentity) + DTO murni + ErrAccountNotFound — SATU-SATUNYA yang boleh diimpor module lain
│   │   ├── domain/{user,role,verification}/{constant,entity,repository,valueobject[,service]}/*.go   # pindahan 1:1 dari internal/domain/{user,role,verification}
│   │   ├── app/
│   │   │   ├── dto/{auth_dto,profile_dto,session_dto,role_permission_dto,user_management_dto}.go
│   │   │   ├── port/{user_query_model,role_permission_query_model}.go   # read-model port milik auth
│   │   │   ├── media/object_path.go                # ObjectPathAvatar + AvatarPresignExpiry
│   │   │   ├── service/
│   │   │   │   ├── principal/builder.go            # implements authz.Resolver
│   │   │   │   └── userdirectory/service.go        # NEW: implements api.UserDirectory — provisioning logic dipindah dari create_santri.go/approve_santri_request.go ke sini
│   │   │   └── usecase/
│   │   │       ├── auth/            (23 file lama + NEW dependencies.go)   # package authusecase
│   │   │       ├── rolepermission/  (19 file, tetap)
│   │   │       └── usermanagement/  (8 file, tetap)
│   │   ├── infrastructure/
│   │   │   ├── persistence/         (7 file postgres_{user,role,verification}_*.go)
│   │   │   └── seed/{role_seeder.go,user_seeder.go}   # dari internal/seeders
│   │   ├── interfaces/http/handler/web/
│   │   │   ├── auth_handler.go, role_permission_handler.go, user_management_handler.go
│   │   │   ├── handlers.go (NEW), routes.go (NEW: RegisterRoutes, isi dari router.go baris auth/users/role-permission)
│   │   │   └── *_test.go (pindahan) — package web_test
│   │   └── testsupport/{seed.go,fixtures.go}       # dari internal/testhelper/{seed,fixtures}.go — package testsupport
│   │
│   └── kesantrian/
│       ├── module.go                               # package kesantrian: Deps{..., UserDirectory kport.UserDirectory}, Module, New, RegisterRoutes
│       ├── domain/santri/{constant,entity,repository,valueobject}/*.go
│       ├── app/
│       │   ├── dto/santri_dto.go                   # dari usecase/santri/dto.go (lihat catatan opsional di bawah)
│       │   ├── media/object_path.go                # ObjectPathSantriDokumen + 2 TTL
│       │   ├── port/user_directory.go              # NEW: port kesantrian sendiri (UserDirectory interface + DTO mirror + ErrAccountNotFound) — INI yang dipakai usecase santri, BUKAN auth/api langsung
│       │   └── usecase/santri/ (17 file, tanpa create_santri's user-building logic — lihat Fase kritis di bawah)
│       ├── infrastructure/
│       │   ├── persistence/ (3 file postgres_santri_*.go — sekalian jadi tx-aware via database.ExecFromContext)
│       │   └── authgateway/user_directory.go       # NEW: satu-satunya file kesantrian yang import modules/auth/api — adapter api.UserDirectory → kesantrian's port.UserDirectory
│       ├── interfaces/http/handler/web/
│       │   ├── santri_handler.go, handlers.go (NEW), routes.go (NEW)
│       │   └── main_test.go, testutil_test.go, santri_handler_test.go (NEW — menutup celah CLAUDE.md §11)
│       └── testsupport/seed.go (opsional, boleh ditambah belakangan)
│
└── testhelper/                                     # TETAP di sini, SATU shared harness (sesuai keputusan user)
    ├── testserver.go     # rewrite: panggil authmodule.New + kesantrianmodule.New + router.New
    ├── testdb.go, noop.go (pakai authz.Principal), txhelper.go (pakai platform/database), assert.go (pakai shared/domain/domainerr)
    └── http.go (NEW: promosi mustRegisterUser/mustLogin/mustSetupSuperadmin dari handler/web/testutil_test.go lama, dipakai kedua module)
```

**Dihapus di akhir** (harus kosong): `internal/app/`, `internal/domain/`, `internal/infrastructure/`, `internal/interfaces/`, `internal/config/`, `internal/logger/`, `internal/seeders/` (termasuk folder kosong untracked `internal/seeders/sources/`).

### Aturan dependency (masuk ke CLAUDE.md, lihat Fase 6)

```
modules/<m>/interfaces  → modules/<m>/app, modules/<m>/domain, shared/*, platform/http/*
modules/<m>/app         → modules/<m>/domain, shared/*, modules/<m>/app/port   (TIDAK BOLEH platform/*, TIDAK BOLEH module lain)
modules/<m>/domain      → shared/domain/domainerr SAJA
modules/<m>/infrastructure → modules/<m>/domain, modules/<m>/app/port, shared/*, platform/database
platform/*              → shared/* saja, TIDAK PERNAH modules/**
shared/*                → stdlib + shared/* saja (leaf)
module A → module B     → HANYA lewat modules/B/api (+ pengecualian terdokumentasi: auth/domain/role/constant untuk permission key di route guard)
adapter lintas-module    → tinggal di infrastructure/<other>gateway/ milik CONSUMER, bukan milik pemilik
```

---

## Cross-module dependency: kesantrian → auth (User)

Sesuai keputusan user (ACL/port lokal), pola konkretnya:

1. **`modules/auth/api/user_directory.go`** — auth menerbitkan kontrak publik (DTO murni: `Account`, `ProvisionAccountInput`, `ProvisionAccountResult`, interface `UserDirectory` dengan `FindAccount`, `FindAccounts` (batch — menghapus N+1), `UpdateFullname`, `ProvisionAccount`, `AttachNISIdentity`, plus `ErrAccountNotFound`).
2. **`modules/auth/app/service/userdirectory/service.go`** — implementasi nyata, depend ke `userrepo.UserRepository` + `port.PasswordHasher` + `port.Transactor`. **Semua logic pembuatan user/credential/login-identity yang saat ini ada di `create_santri.go` (baris pembuatan `userentity.User`, 3x `NewLoginIdentity`, `generateRandomPassword`, hash) dan `approve_santri_request.go` (attach NIS identity) PINDAH ke sini.**
3. **`modules/kesantrian/app/port/user_directory.go`** — port versi kesantrian sendiri (DTO mirror + interface sama bentuknya), inilah yang di-inject ke usecase santri (`Dependencies`).
4. **`modules/kesantrian/infrastructure/authgateway/user_directory.go`** — SATU-SATUNYA file di kesantrian yang import `modules/auth/api`; adapter struct-to-struct + error mapping. Di-wire di `cmd/app/main.go`: `authgateway.New(authMod.UserDirectory)`.

Efek pada usecase santri yang harus disesuaikan (bukan sekadar pindah file, ini perubahan semantik — beri commit terpisah):
- `Dependencies` kesantrian **kehilangan** `Hasher` (`port.PasswordHasher`) — hashing sekarang tanggung jawab auth.
- `CreateSantriResponse.PasswordGenerated` diisi dari `ProvisionAccountResult.GeneratedPassword`.
- `create_santri.go` + `santriRepo.Save` idealnya jadi **atomik** lewat `Transactor.WithTx` (saat ini TIDAK atomik — kalau `santriRepo.Save` gagal, user/credential/login-identity yang sudah tersimpan jadi orphan). Karena santri repo saat ini belum tx-aware (tidak pernah panggil `execFromContext`), sekalian jadikan tx-aware saat dipindah ke `database.ExecFromContext`.
- `list_santri.go` dan `list_santri_requests.go` ganti dari lookup user satu-satu dalam loop menjadi panggil `FindAccounts` (batch).

---

## Urutan eksekusi (per fase, build harus tetap hijau di akhir tiap fase)

Alasan urutan: shared/platform dulu (semua bergantung ke sini, dan ini satu-satunya fase yang tidak bisa dicicil), lalu tiap module bottom-up (domain → persistence → app → interfaces), baru rewire router+main+testhelper di akhir.

**Command validasi tiap fase:**
```bash
gofmt -l ./cmd ./internal   # harus kosong
go build ./... && go vet ./...
go test ./... -count=1 -timeout 300s
```
Pakai `git mv` (bukan delete+recreate) supaya rename terdeteksi git, lalu `sed` untuk ganti import path. Satu commit per fase (Fase 5 dipecah jadi 3 commit: pindah file / userdirectory-refactor / dto-extraction).

1. **Fase 1 — shared + platform** (±150 file, satu commit besar tak terhindarkan karena Go tidak bisa split package secara bertahap). Export `execFromContext`→`ExecFromContext`, `resolvePaginationParams`→`ResolvePaginationParams`, `nullableString`→`NullableString`. Semua import path di seluruh repo diupdate. Module code (domain/usecase/dll) TETAP di lokasi lama fase ini — hanya dependency-nya yang pindah.
2. **Fase 2 — auth: domain + persistence + seeder.** Pindah `domain/{user,role,verification}` dan `postgres_{user,role,verification}_*.go` + `seeders/{role,user}_seeder.go` ke `modules/auth/...`. Hapus `CodeInvalidNISFormat` duplikat yang dead code.
3. **Fase 3 — auth: app layer.** Pindah dto, port (user_query_model, role_permission_query_model), `service/principal/builder.go`, ke-23 file usecase `auth`, 19 file `rolepermission`, 8 file `usermanagement`. Tambah `dependencies.go` untuk `auth` usecase (belum ada). Tambah (additive, belum dipakai) `api/user_directory.go` + `service/userdirectory/service.go`.
4. **Fase 4 — auth: interfaces + module.go.** Pindah 3 handler + test-nya, tambah `handlers.go`/`routes.go`/`module.go`. Router lama tetap monolitik sampai Fase 6, jadi masih hijau.
5. **Fase 5 — kesantrian module** (3 commit terpisah):
   - **5a (mekanis):** pindah domain santri, persistence santri (sekalian jadi tx-aware), 17 file usecase, handler, tambah `handlers.go`/`routes.go`/`module.go`.
   - **5b (semantik):** refactor usecase santri lepas dari `userrepo` langsung → pakai `port.UserDirectory` (lihat bagian Cross-module di atas); hapus `Hasher` dari `Dependencies`; pindah `generateRandomPassword` + logic pembuatan user ke `userdirectory.Service`; ganti list usecase ke `FindAccounts` batch.
   - **5c (opsional, mekanis tapi luas):** ekstrak `usecase/santri/dto.go` → `app/dto/santri_dto.go` supaya konsisten dengan konvensi dto module lain (CLAUDE.md §2/§8 mengasumsikan dto terpisah). Kalau mepet waktu, boleh ditunda — tapi catat sebagai deviasi terdokumentasi di CLAUDE.md.
6. **Fase 6 — router kernel, main.go, seeder CLI, testhelper, tooling:**
   - Rewrite `platform/http/router/router.go` jadi `Kernel`/`Groups`/`RouteRegistrar` (body route pindah ke masing-masing module `routes.go`).
   - Rewrite `cmd/app/main.go` (~240 baris → target ≤130): bangun `authMod := authmodule.New(...)`, `kesantrianMod := kesantrianmodule.New(Deps{..., UserDirectory: authgateway.New(authMod.UserDirectory)})`, lalu `router.New(Kernel{...}, authMod, kesantrianMod)`.
   - `cmd/seeder/main.go`: seeder list sekarang eksplisit dari sisi caller (`authseed.RoleSeeder{}, authseed.UserSeeder{}`), karena `platform/seeder` tidak boleh tahu soal module.
   - `internal/testhelper/testserver.go`: rewrite untuk wiring lewat `authmodule.New` + `kesantrianmodule.New` + `router.New` (tetap SATU shared TestServer sesuai keputusan user). Tambah `testhelper/http.go` (promosi helper dari `testutil_test.go` lama). Buat `modules/kesantrian/interfaces/http/handler/web/{main_test.go,testutil_test.go,santri_handler_test.go}` baru (menutup celah §11: 401 unauth, 422 NIS invalid, 404 not-found, 403 non-admin di `/santri/admin/*`, 2xx happy path).
   - Update `Makefile` target test (`test-unit`, `test-integration`, `test-usecase`, tambah `test-handler`) supaya path-nya ikut struktur baru — kalau tidak, target-target itu diam-diam jadi no-op.
   - `make swagger` regenerate ulang.
7. **Fase 7 — cleanup:** hapus direktori lama yang sudah kosong, tambah `var _ router.RouteRegistrar = (*authmodule.Module)(nil)` dkk sebagai compile-time check, update `.claude/CLAUDE.md` (lihat bawah).

Setelah Fase 1/2/3/5b, jalankan `git diff --stat` untuk sanity-check bahwa fase murni-pindah tidak mengubah baris di luar blok import.

**Peringatan konflik nama package saat pindah:** `dto`, `port`, `media`, `web`, `persistence` sekarang eksis di lebih dari satu tempat (shared vs per-module). JANGAN jalankan `goimports` membabi-buta lintas repo di fase yang punya package sama nama — bisa salah pilih import (contoh riskan: `dto.Meta` shared vs `dto.Meta` milik module, sama-sama valid secara compile tapi salah semantik). Tambahkan alias secara manual di ~15 file yang butuh keduanya (pola: `shareddto`, `sharedport`, `sharedmedia` untuk yang shared; nama polos untuk yang module-lokal), lalu `gofmt`.

---

## Update CLAUDE.md (`.claude/CLAUDE.md`, bukan root)

Tambahkan/ubah:
- §1: path jadi relatif-module (`internal/modules/<module>/interfaces/...`, dst).
- **§1.5 baru "Struktur Modular Monolith"**: tree ringkas + daftar module (`auth`, `kesantrian`) + aturan kapan sesuatu masuk `shared/` vs `platform/` vs `modules/`.
- **§1.6 baru "Aturan Dependensi Antar Modul"**: tabel dependency di atas, termasuk aturan "module lain hanya lewat `api` package" dan pengecualian permission-constant.
- §3: diagram dependency versi per-module.
- §5/§5.1: path repository/port kontrak sekarang di bawah module; §5.1 pattern filename tetap sama, cuma pindah lokasi.
- **§5.2 baru "Package Naming & Alias Convention"**: tabel alias (`shareddto`/`sharedport`/`sharedmedia`, `authweb`/`kesantrianweb`, dst) supaya konsisten ke depan.
- §8: path pagination dto pindah ke `internal/shared/app/dto/pagination_dto.go`.
- §9: ObjectPath sekarang split: tipe di shared, konstanta per module.
- §11: update tree testhelper + contoh command `go test`, tegaskan aturan "satu shared harness, instance per-module test package".
- §12: tambah satu kalimat bahwa migrasi sengaja tetap flat, tidak ikut dipecah per module.
- **§14 baru "Menambah Modul Baru"**: checklist singkat untuk module berikutnya (bikin 4-layer folder, `module.go`, `handlers.go`/`routes.go`, daftar di `main.go` + `testhelper`, define cross-module need via port+adapter).

---

## Berkas kritis untuk implementasi

- `/home/nurdiansyah/Desktop/sipon/sipon-api/cmd/app/main.go` — wiring monolitik yang di-rewrite jadi per-module.
- `/home/nurdiansyah/Desktop/sipon/sipon-api/internal/interfaces/http/router/router.go` — sudah dibaca; jadi basis `Kernel`/`Groups`/`RouteRegistrar` + isi tiap module `routes.go`.
- `/home/nurdiansyah/Desktop/sipon/sipon-api/internal/infrastructure/persistence/postgres_transactor.go` — sudah dibaca; export `Executor`/`ExecFromContext` jadi prasyarat Fase 1.
- `/home/nurdiansyah/Desktop/sipon/sipon-api/internal/app/usecase/santri/create_santri.go` — sudah dibaca; sumber logic yang harus pindah ke `userdirectory.Service` di Fase 5b.
- `/home/nurdiansyah/Desktop/sipon/sipon-api/internal/testhelper/testserver.go` — basis rewrite Fase 6.
- `/home/nurdiansyah/Desktop/sipon/sipon-api/.claude/CLAUDE.md` — diupdate di Fase 7.

---

## Verifikasi

1. Setiap fase: `gofmt -l ./cmd ./internal` (kosong), `go build ./... && go vet ./...`, `go test ./... -count=1 -timeout 300s` — semua hijau sebelum lanjut fase berikutnya.
2. Setelah Fase 6: jalankan `make dev-up && make run`, smoke test manual — `GET /health`, register+login lewat `/api/v1/web/auth/*`, `GET /api/v1/web/santri/profile` dengan token. Bandingkan daftar route (`gin` debug route log) sebelum vs sesudah restructure untuk memastikan tidak ada route yang hilang saat `router.go` dipecah.
3. Cek `Makefile` test target (`test-unit`, `test-integration`, `test-usecase`, `test-handler`) benar-benar menjalankan test dan tidak diam-diam kosong (`go test <path> -v` harus menunjukkan test dijalankan, bukan "no test files").
4. `make swagger` dijalankan ulang, diff `docs/swagger.json` dipastikan endpoint tidak hilang.
5. Review manual khusus untuk 5 file yang secara semantik berubah (bukan cuma pindah): `platform/http/middleware/auth.go`, `platform/database/transactor.go`, `platform/database/pagination.go`, `shared/authz/cache.go`, `platform/seeder/seeder.go`, plus `modules/auth/app/service/userdirectory/service.go` (logic baru) dan santri usecase yang direfactor (Fase 5b).
