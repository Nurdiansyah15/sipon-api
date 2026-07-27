# Dark Mode — Penerapan di Seluruh Halaman sipon-ui

Status: implemented.
Scope: `sipon-ui` (frontend only — backend tidak terpengaruh).

## Context

sipon-ui menggunakan Nuxt UI v4 (`@nuxt/ui` v4.10.0) yang memiliki **built-in color mode
support**. Semua komponen `U*` (UButton, UCard, UTable, UBadge, dsb.) sudah otomatis beradaptasi
antara light/dark mode melalui CSS variables yang di-generate oleh Nuxt UI.

Namun, **tidak ada satupun halaman/komponen** yang menerapkan `dark:` variant pada kelas Tailwind
kustom. Akibatnya:

- Halaman terang-terangan hanya mendukung light mode (`bg-white`, `text-gray-900`, `border-gray-200` dsb.)
- Toggle "Dark Mode" di `AppUserMenu.vue` adalah placeholder — mengubah `ref` lokal tanpa efek apapun
- Tidak ada konfigurasi `colorMode` di `app.config.ts` atau `nuxt.config.ts`

### Current state

| Item | Status |
|---|---|
| Nuxt UI v4 color mode support | Built-in, tidak perlu modul tambahan |
| `app.config.ts` — `ui.colors` | Sudah diset `primary: 'teal'`, `neutral: 'slate'` |
| `app.config.ts` — `ui.colorMode` | **Belum ada** |
| `app.vue` — `<UColorModeScript>` | **Belum ada** |
| `AppUserMenu.vue` — dark mode toggle | Placeholder, tidak terhubung ke `useColorMode()` |
| `dark:` variant di komponen/pages | **0 file** dari ~28 file |
| Layout & auth pages | Semua hardcoded light-mode colors |

---

## Plan

### Phase 1 — Enable color mode infrastructure (4 file)

**1.1 `app.config.ts`** — Tambah konfigurasi `colorMode`:

```ts
export default defineAppConfig({
  ui: {
    colors: {
      primary: 'teal',
      neutral: 'slate',
    },
    colorMode: {
      preference: 'system',
      fallback: 'light',
    },
  },
})
```

**1.2 `app.vue`** — Tambah `<UColorModeScript>` untuk mencegah flash of wrong theme saat SSR:

```vue
<template>
  <UApp>
    <UColorModeScript />
    <NuxtLoadingIndicator />
    ...
  </UApp>
</template>
```

**1.3 `AppUserMenu.vue`** — Wire toggle ke `useColorMode()`:

Ganti `const darkMode = ref(false)` dengan composable `useColorMode()`.

**1.4 `main.css`** — Opsional: tambah `color-scheme: light dark` untuk transisi smooth.

### Phase 2 — Apply `dark:` variants ke semua halaman & komponen (~28 file)

#### Color Mapping Table

| Light class | Dark variant |
|---|---|
| `bg-white` | `dark:bg-gray-900` |
| `bg-gray-50` | `dark:bg-gray-950` |
| `bg-gray-100` | `dark:bg-gray-800` |
| `bg-gray-200` | `dark:bg-gray-700` |
| `text-gray-900` | `dark:text-gray-100` |
| `text-gray-800` | `dark:text-gray-200` |
| `text-gray-700` | `dark:text-gray-300` |
| `text-gray-600` | `dark:text-gray-400` |
| `text-gray-500` | `dark:text-gray-400` |
| `text-gray-400` | `dark:text-gray-500` |
| `border-gray-200` | `dark:border-gray-700/50` |
| `border-gray-300` | `dark:border-gray-600/50` |
| `bg-teal-50` | `dark:bg-teal-950` |
| `text-teal-600` | `dark:text-teal-400` |

**Prinsip**: Hanya kelas Tailwind kustom yang disentuh. Komponen Nuxt UI sudah auto.

#### Layouts (2 file)

- `layouts/default.vue` — `bg-gray-50`
- `layouts/system-admin.vue` — `bg-gray-50`

#### Pages (8 file)

- `pages/auth/login.vue`, `pages/auth/register.vue` — card, form fields, text, divider
- `pages/dashboard/index.vue` — judul
- `pages/profile/index.vue` — panel, border, teks
- `pages/system-admin/index.vue` — judul, deskripsi
- `pages/system-admin/users/index.vue` — tabel, filter, tombol, pagination
- `pages/system-admin/roles/index.vue` — tabel, filter, tombol
- `pages/system-admin/roles/[id]/permissions.vue` — list permission, toggle

#### Components (18 file)

Global: `AppNavbar`, `AppFooter`, `AppMobileBottomNav`, `AppSystemAdminNavbar`, `AppSystemAdminMobileNav`, `AppRowActions`
Dashboard: `HeroBanner`, `FeatureModuleCard`, `FeatureModuleGrid`
System-admin: `CreateUserModal`, `CreateRoleModal`, `AssignRoleModal`, `ResetPasswordResultModal`, `OneTimePasswordReveal`
Profile: `AccountInfoPanel`, `SetPasswordForm`, `ChangePasswordForm`, `RolesPermissionsPanel`

---

## Verification

1. Buka `/dashboard` — background, card, teks terbaca di dark mode
2. Buka `/auth/login`, `/auth/register` — branded bg tetap terlihat, card jelas
3. Buka `/profile` — semua panel terbaca
4. Buka `/system-admin/*` — semua halaman admin
5. Cek mobile viewport — navbar, bottom nav di dark mode
6. Refresh halaman — tidak ada flash of wrong theme
7. Toggle dark mode via user menu — pastikan bekerja

## Critical files

- `app/app.config.ts` — tambah `ui.colorMode`
- `app/app.vue` — tambah `<UColorModeScript>`
- `app/components/AppUserMenu.vue` — wire `useColorMode()`
- `app/assets/css/main.css` — tambah CSS kustom
- Seluruh 28 file .vue (layouts, pages, components) — tambah `dark:` variants
