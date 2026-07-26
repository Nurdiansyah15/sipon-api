# sipon-api

Backend API untuk aplikasi Sipon. Dibangun dengan Go, menggunakan PostgreSQL dan Redis.

## Prasyarat

- Docker & Docker Compose
- Go 1.22+ (untuk development tools)
- Make

## Setup Awal

```bash
# 1. Clone & masuk direktori
cd sipon-api

# 2. Copy environment file (sudah ada default untuk development)
cp .env.example .env
# Edit .env jika perlu, terutama JWT_SECRET_KEY

# 3. Jalankan PostgreSQL + Redis (tunggu sampai healthy)
make dev-up

# 4. Jalankan migrasi database
make migrate-up

# 5. (Opsional) Seed role dan data awal
make seed-role

# 6. Jalankan server
make run
```

Server akan berjalan di `http://localhost:8800`.

## Perintah Penting

| Perintah | Fungsi |
|----------|--------|
| `make dev-up` | Jalankan PostgreSQL + Redis |
| `make dev-down` | Hentikan container dev |
| `make run` | Jalankan HTTP server |
| `make migrate-up` | Jalankan migrasi |
| `make migrate-down` | Rollback migrasi terakhir |
| `make migrate-fresh` | Reset DB + migrasi ulang |
| `make seed-role` | Seed role default |
| `make seed-all` | Jalankan semua seeder |
| `make test` | Jalankan semua test |
| `make swagger` | Generate dokumentasi API |

## Port Mapping (Development)

| Service | Internal Port | Exposed Port |
|---------|--------------|--------------|
| App | 8800 | 8800 |
| PostgreSQL | 5432 | 5435 |
| Redis | 6379 | 6381 |

## Struktur Proyek

```
sipon-api/
├── cmd/            # Entry point
├── internal/       # Kode aplikasi (handler, usecase, repository, dll)
│   ├── migrations/ # File migrasi database
│   └── ...
├── docs/           # Dokumentasi Swagger (generated)
├── docker-compose.dev.yml
├── Makefile
└── .env.example
```

## API Docs

Setelah server berjalan, akses dokumentasi API di `http://localhost:8800/swagger/index.html`.
Generate ulang dengan `make swagger` jika ada perubahan handler.
