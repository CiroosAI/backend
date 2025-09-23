# SETUP-LOCAL (Developer Test Guide)

This document explains how to run the application locally using Docker (recommended) or directly from your machine for quick tests.

Prerequisites
- Docker Desktop (Windows) or Docker + docker-compose
- Go 1.20+ (if running locally)
- Git

1) Start the development stack (Docker Compose)

From project root:
```powershell
# Build and start MySQL (loaded from database/full_db.sql), Redis and the app
docker compose -f docker-compose.dev.yml up -d --build

# View logs
docker compose -f docker-compose.dev.yml logs -f app
```

2) Verify services
```powershell
docker ps
# Check Redis is healthy
docker exec -it sf-redis-dev redis-cli ping
```

3) Run the app locally (optional)
- If you want to run Go app on host and use container DB/Redis, export env values (PowerShell example):
```powershell
$env:ENV="development"
$env:DB_HOST="127.0.0.1"
$env:DB_PORT="3307" # mapped in compose.dev
$env:DB_USER="root"
$env:DB_PASS="rootpassword"
$env:DB_NAME="sf"
$env:JWT_SECRET="supersecretjwtkey"
$env:REDIS_ADDR="127.0.0.1:6379"

go run main.go
```

4) Test login brute-force
- Use the `/login` endpoint and attempt wrong password. With Redis enabled, account lockout will be enforced and persisted across restarts. IP rate-limiter is also active.

5) Tear down
```powershell
docker compose -f docker-compose.dev.yml down -v
```

Notes
- The `database/full_db.sql` is loaded into MySQL when container initializes. If you change the DB schema, rebuild containers or migrate using a migration tool.
- Do not commit any secrets to the repo. Use `.env.local` for local secrets and `.env` for production overrides (managed on VPS).

powershell -NoProfile -ExecutionPolicy Bypass -File "c:\Users\USER\Website Client\StoneForm\BackEnd-LocalHost\scripts\run-local.ps1"
