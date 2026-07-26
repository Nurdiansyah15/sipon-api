# System Management Module — Permissions, User Management, Role/Permission UI

Status: planned, not yet implemented.
Scope: `sipon-api` (this repo) + `sipon-ui` (sibling frontend).

## Context

Sipon needs a "System Management" application — a module tile reachable from the dashboard's
"Aplikasi" grid in sipon-ui — so admins can:

- **Manage users**: create user, reset password, activate/deactivate user, assign/revoke user roles.
- **Manage roles & permissions**: create role, edit role, edit a role's permission set.

Current state before this feature:

- Role/permission backend is mostly complete (`/api/v1/web/role-permission/*`: role CRUD,
  permission assignment, user-role assign/revoke) but only 3 permission keys exist
  (`manage_system_settings`, `assign_role`, `manage_users`), and the whole `/role-permission/*`
  group is guarded by a blunt `middleware.RequireRole("superadmin","usergod")` even though a
  generic `middleware.RequirePermission(...)` already exists and goes unused. This is a **latent
  bug**: `admin`'s `RolePermissions` entry already includes `assign_role`, but the role-check
  blocks `admin` from ever reaching the route that would use it.
- No admin-facing user management endpoints exist at all — only self-service
  register/login/forgot-password/change-password/change-identity under `internal/app/usecase/auth/`.
- sipon-ui already has a navbar-based dashboard template (`AppNavbar`, `AppMobileBottomNav`,
  `AppFooter`, `FeatureModuleGrid`/`FeatureModuleCard`, `HeroBanner`) but the "Aplikasi" grid cards
  are purely decorative (no navigation yet), and there's no RBAC scaffolding
  (`usePermission`/`PermissionGate`/permission middleware) or CRUD page pattern yet. A sibling
  reference project `k-forum-backoffice` has the RBAC + CRUD-page patterns to adapt (its layout is
  sidebar-based; sipon-ui's is navbar-based — borrow composables/components, not the outer
  dashboard-panel wrapper).

### Decisions already made

1. Admin-created users get an **auto-generated password shown once** in the API response (no
   email infrastructure exists in sipon-api today) — the frontend must present it with a copy
   button and a clear "won't be shown again" warning.
2. **Deactivate/activate reuses the existing `ACTIVE`/`BANNED` user status** — no migration for a
   new status value.
3. The existing `/role-permission/*` routes will be **retrofitted to use granular
   `RequirePermission`** instead of the blanket `RequireRole`, fixing the admin-lockout bug as a
   side effect.
4. `DELETE /roles/:role_id` is **out of scope** (not requested; touches referential-integrity
   concerns — existing `user_roles`/`role_permissions` rows — unrelated to this feature).
5. "Force change password on first login" is **explicitly deferred** — would need a new column, a
   migration, and a login-flow gate; noted as a follow-up ticket, not built now.

---

## Backend plan (sipon-api)

Follow `.claude/CLAUDE.md` conventions throughout: one usecase per file, `respond.SuccessBody`/
`ErrorBody` envelope, migrations via `make migrate-create NAME=...`, repository error mapping
(`sql.ErrNoRows`→NotFound, insert/update/delete failure→PersistenceFailed, query failure→
QueryFailed, pg `23505`→DuplicateKey), mandatory handler tests covering 401/400/404/403/2xx.

### 1. Permission constants

File: `internal/domain/role/constant/permission_constant.go`

Keep existing keys unchanged (`manage_system_settings`, `assign_role`, `manage_users`) — no
renames, avoids breaking stored custom-role `role_permissions` rows. Add:

```go
PermissionResetUserPassword     PermissionKey = "reset_user_password"     // NEW
PermissionDeactivateUser        PermissionKey = "deactivate_user"         // NEW — covers deactivate + reactivate
PermissionManageRoles           PermissionKey = "manage_roles"            // NEW — create/edit role metadata
PermissionManageRolePermissions PermissionKey = "manage_role_permissions" // NEW — assign/revoke permission on a role
```

Update `DefaultPermissionsInit` with display name/description for each new key, and `RolePermissions`:

| Role | Permissions |
|---|---|
| `usergod` | all 7 keys |
| `superadmin` | all 7 keys |
| `admin` | `assign_role`, `manage_users`, `reset_user_password`, `deactivate_user` (NOT `manage_roles`/`manage_role_permissions` — role/permission *definitions* stay superadmin/usergod-only) |
| `member` | none |

No changes needed to `internal/seeders/role_seeder.go` (reads `DefaultRolesInit`, not
`RolePermissions` — the latter is read live by `principal.Builder`).

### 2. Migrations

**None required.** Confirmed: user status reuses `ACTIVE`/`BANNED` (existing CHECK constraint in
`internal/migrations/20260726120000_create_auth_rbac_tables.up.sql`), password reset reuses the
existing `credentials.secret_hash` column.

### 3. Domain — `User` entity

File: `internal/domain/user/entity/user.go` — add, next to the existing `Activate()`:

```go
func (u *User) Deactivate() error { // -> StatusBanned, guards against double-deactivate
func (u *User) Reactivate() error { // -> StatusActive, guards against double-reactivate
```

New domain error codes (wherever `CodeUser...` constants live): `CodeUserAlreadyBanned`,
`CodeUserAlreadyActive`, `CodeUserDuplicate` (see §4.1).

### 4. Persistence

**4.1 Fix missing duplicate-key mapping** — `internal/infrastructure/persistence/postgres_user_repository.go`'s
`Save()` doesn't currently map Postgres `23505` to a domain duplicate error (unlike
`postgres_role_permission_repository.go`). Add it (`CodeUserDuplicate`) since admin-create-user
calls `Save` directly with an admin-supplied email/username/phone, and the pre-insert existence
checks are a TOCTOU race — the DB constraint is the real backstop.

**4.2 New read model for admin user listing** — `internal/app/port/user_query_model.go`
(`UserListReadQuery`, `UserReadItem`, `UserQueryReadModel` interface) +
`internal/infrastructure/persistence/postgres_user_query.go` (`PostgresUserQuery.ListUsers`).
Filters: `status`, `role_id`, `search` (ILIKE username/fullname/email). `role_names` via a
correlated subquery (not a JOIN, to avoid multiplying rows and breaking pagination). Uses the
same `resolvePaginationParams` helper as `postgres_role_query.go`.

The domain `UserRepository` interface stays untouched — listing belongs in the read-model/query
layer per CLAUDE.md §6, not the repo interface (mirrors why `RoleRepository` has no `ListRoles`).

### 5. Usecases

New package `internal/app/usecase/usermanagement/` (separate from `rolepermission` — operates on
the `user` aggregate with different dependencies: `PasswordHasher`, no role-permission repo).
One file per usecase, `Dependencies`/`UseCases`/`NewUseCases` factory in `dependencies.go`, shared
helpers in `helpers.go`:

- `mapUserDomainError` — maps `CodeUserNotFound`→404, `CodeUserDuplicate`→409,
  `CodeUserAlreadyBanned`/`CodeUserAlreadyActive`→409, persistence/query failures→500.
- `generateTemporaryPassword()` — `crypto/rand`, satisfies `valueobject.NewPlainPassword`'s rules
  (min 8 chars, ≥1 uppercase, ≥1 digit), re-validated defensively before hashing.

Usecases:
- `list_users.go` — perm `manage_users`, paginated, `(data, meta, error)` per CLAUDE.md §8.
- `get_user.go` — perm `manage_users`, includes active role assignments.
- `create_user.go` — perm `manage_users`, builds `User`+`Credential`+`LoginIdentity` like
  `RegisterUseCase` but skips OTP/verification and auto-role-assignment (admin assigns roles
  afterward via the existing `user-roles` endpoint). Returns `generated_password`.
- `reset_user_password.go` — perm `reset_user_password`, generates+hashes a new temp password,
  resets failed-login attempts.
- `deactivate_user.go` / `reactivate_user.go` — perm `deactivate_user`.

### 6. DTOs

File: `internal/app/dto/user_management_dto.go` — `ListUsersQuery`, `UserManagementResponse`
(`roles` populated only on get-by-id, omitted on list to avoid N+1), `CreateUserRequest`,
`CreateUserResponse` (embeds `generated_password` — shown once, never retrievable again),
`ResetUserPasswordResponse`.

### 7. Handler + routes

File: `internal/interfaces/http/handler/web/user_management_handler.go` (mirrors
`role_permission_handler.go`'s style exactly — swaggo blocks, `ShouldBindJSON`/`ShouldBindQuery`,
`respond.OK`/`Created`/`SuccessWithMeta`).

New route group in `internal/interfaces/http/router/router.go`, under `/api/v1/web/users`:

| Route | Method | Guard |
|---|---|---|
| `/users` | GET | `RequirePermission(manage_users)` |
| `/users/:user_id` | GET | `RequirePermission(manage_users)` |
| `/users` | POST | `RequirePermission(manage_users)` |
| `/users/:user_id/reset-password` | POST | `RequirePermission(reset_user_password)` |
| `/users/:user_id/deactivate` | POST | `RequirePermission(deactivate_user)` |
| `/users/:user_id/reactivate` | POST | `RequirePermission(deactivate_user)` |

**Retrofit `/role-permission/*`** — replace the blanket `RequireRole("superadmin","usergod")`
group guard with per-route `RequirePermission`:

| Route(s) | Guard |
|---|---|
| `GET /roles`, `/roles/:id`, `/permission-keys` | `RequirePermission(manage_roles, manage_role_permissions, assign_role)` (any grants read visibility) |
| `POST/PUT /roles*` | `RequirePermission(manage_roles)` |
| `POST/DELETE /roles/:id/permissions*` | `RequirePermission(manage_role_permissions)` |
| `GET /user-roles*` | `RequirePermission(assign_role, manage_users)` |
| `POST/PUT/DELETE /user-roles*` (assign/update/deactivate/reactivate/revoke) | `RequirePermission(assign_role)` |

### 8. DI wiring

`cmd/app/main.go` — construct `PostgresUserQuery`, `usermanagement.NewUseCases(...)`,
`webhandler.NewUserManagementHandler(...)`, pass into `router.Setup(...)`.

### 9. Tests

New `internal/interfaces/http/handler/web/user_management_handler_test.go` (package `web_test`,
reuse `testSrv`/`superadminHeader()`/helpers from `main_test.go`/`testutil_test.go`):
401 (no auth), 403 (member), 400/422 (missing fields), 404 (bogus id), 2xx (full create → get →
reset-password → deactivate → reactivate happy path, asserting `generated_password` present and
changes between create/reset). Plus a regression case: an `admin`-role user gets 200/201 on
`manage_users`-guarded routes and on `assign_role`-guarded `/user-roles` routes (proving the
lockout bug is fixed), but 403 on `manage_roles`-guarded `POST /roles` (proving admin still lacks
role-definition rights). Update `role_permission_handler_test.go` accordingly.

Run: `go test ./internal/interfaces/http/handler/web/... -count=1 -timeout 300s`.

---

## Frontend plan (sipon-ui)

### RBAC scaffolding (build first)

Adapted from `k-forum-backoffice`, simplified to sipon-ui's actual `authStore.hasRole`/
`hasPermission` API (no plan/region concepts):

- `app/composables/usePermission.ts` — `can(key)` / `canAny(keys)` / `canAll(keys)` / `hasRole(role)`.
- `app/components/PermissionGate.vue` — `permission?`/`permissions?`/`requireAll?`/`role?` props, `#fallback` slot.
- `app/middleware/permission.ts` — named middleware reading `to.meta.permission = { permissions?, roles?, requireAll? }`, redirects to `/forbidden`.
- `app/pages/forbidden.vue` — simple denial page, default layout.

### Navigation wiring

- `FeatureModuleCard.vue` — add optional `to` prop (renders as link when present; existing 6
  decorative modules untouched).
- `FeatureModuleGrid.vue` — append a "Manajemen Sistem" tile → `/system-management`, visible only
  when `canAny(['manage_users','manage_roles','manage_role_permissions'])`.
- `AppNavbar.vue`/`AppMobileBottomNav.vue` — left untouched; the module is reached via the
  dashboard tile, not a new top-level nav item. No new layout file needed — `layouts/default.vue`
  already provides the navbar (desktop) + bottom icon nav (mobile) for any new page.

### Types & stores

`shared/types/UserManagement.ts`, `shared/types/RolePermission.ts` (verify exact field names
against real sipon-api responses while implementing), `app/stores/userManagement.ts`
(fetchUsers/fetchUser/createUser/resetPassword/deactivateUser/reactivateUser — **never persist the
one-time generated password in store state**), `app/stores/rolePermission.ts`
(fetchRoles/fetchRole/createRole/updateRole/fetchPermissionKeys/assignPermission/
revokePermission/assignUserRole/deactivate-or-revoke user-role).

### Shared UI

`app/components/AppDataTable.vue` (minimal `UTable`+`UPagination` wrapper, no column-visibility/
sorting complexity for v1), `app/components/system-management/OneTimePasswordReveal.vue` (shared
"copy this password, it won't be shown again" body via `useClipboard()` from `@vueuse/core`).

### Pages

All under `app/pages/system-management/`, all inheriting `layouts/default.vue` automatically:

- `system-management.vue` — shell: permission-gated redirect `/system-management` →
  `/system-management/users`, route-driven `UTabs` (Pengguna / Peran & Hak Akses — second tab only
  if `canAny(['manage_roles','manage_role_permissions'])`), `<NuxtPage/>`. Note: `@nuxt/ui` v4's
  `TabsItem` has no `to` field — drive it via `:model-value`/`@update:model-value` → `navigateTo(...)`.
- `system-management/users/index.vue` — search/status filter, paginated `AppDataTable`, "Buat
  Pengguna" → `CreateUserModal`, row actions (Kelola Peran → `AssignRoleModal`; Setel Ulang Kata
  Sandi → `ResetPasswordResultModal`; Nonaktifkan/Aktifkan), each gated per the matching permission.
- `system-management/roles/index.vue` — role list + "Buat Peran" → `CreateRoleModal`, link to
  permissions editor (gated on `manage_role_permissions`).
- `system-management/roles/[id]/permissions.vue` — toggle list of all permission keys (from
  `GET /role-permission/permission-keys`, source of truth, not hardcoded), each toggle immediately
  firing assign/revoke.
- No separate `users/[id].vue` detail page for v1 — actions happen via modals from the table row.

### Forms

Zod + `UForm` for every mutating form (create user, create role, assign role) per CLAUDE.md's
mandatory convention.

### Password-reveal UX

`CreateUserModal`/`ResetPasswordResultModal` must not silently close after success — show the
generated password with a copy button and a clear "won't be shown again" warning; block dismissal
(`:dismissible="false"`) until the admin explicitly confirms they've copied it.

---

## Verification

- Backend: `go test ./internal/interfaces/http/handler/web/... -count=1 -timeout 300s`.
- Frontend: `npm run dev`, log in as seeded superadmin/admin/member, manually verify: dashboard
  tile visibility per role; create user → one-time password shown → new user can log in; reset
  password; deactivate/reactivate; create role → edit permissions → a member assigned that role
  gains/loses access accordingly. Check mobile viewport (bottom nav, responsive tables/forms).

## Critical files

- `internal/domain/role/constant/permission_constant.go`
- `internal/domain/user/entity/user.go`
- `internal/app/usecase/rolepermission/dependencies.go` (pattern for new `usermanagement` package)
- `internal/infrastructure/persistence/postgres_role_query.go` (pattern for `postgres_user_query.go`)
- `internal/interfaces/http/router/router.go`
- `cmd/app/main.go`
- `internal/interfaces/http/handler/web/role_permission_handler_test.go` (pattern + needs the admin-permission regression test added)
- `../sipon-ui/app/stores/auth.ts`
- `../sipon-ui/app/components/FeatureModuleGrid.vue` / `FeatureModuleCard.vue`
- `../sipon-ui/app/layouts/default.vue`
- `/home/nurdiansyah/Desktop/k-forum-backoffice/app/composables/usePermission.ts` + `app/middleware/permission.ts` (pattern reference only)
