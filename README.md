# go-kinerja-backend

Backend untuk penilaian kinerja karyawan (Go + Gin + Postgres + ClickUp).

## Setup singkat

1. copy `.env.example` → `.env` dan isi variabel (terutama CLICKUP_TOKEN, JWT_SECRET)
2. jalankan docker postgres:
   docker compose -f docker/docker-compose.yml up -d
3. jalankan migration:
   psql -h localhost -U postgres -d kinerja_db -f internal/db/migrations/001_init.sql
4. seed admin:
   export ADMIN_PASSWORD=dnakinerja
   go run scripts/seed_admin.go
5. run server:
   go run cmd/server/main.go

Login admin:
POST /api/v1/auth/login
payload:
{ "username": "admin", "password": "dnakinerja" }

