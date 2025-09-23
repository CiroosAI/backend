# SETUP (Production VPS)

This guide outlines steps to deploy the application on a VPS using Docker and Docker Compose. It focuses on production hardening and operational concerns.

Prerequisites
- VPS with Ubuntu 22.04 or similar
- Docker Engine and Docker Compose v2 installed
- A domain name and TLS certificate (Let's Encrypt)
- An SMTP provider or notification channel
- Secrets management (recommended: Vault, AWS Secrets Manager or environment variables via systemd or Docker Swarm/Kubernetes)

High-level steps
1. Create non-root deploy user and install Docker & Compose
2. Pull repository to VPS (git clone)
3. Create production `.env` with secure secrets (never commit)
4. Use `docker-compose.prod.yml` to start services: app, mysql, redis, and a reverse proxy (Caddy or Nginx) with TLS
5. Run DB migrations (use migration tool or manually run SQL)
6. Monitor logs and set up systemd unit or Docker restart policies

Important security hardening
- Do not run AutoMigrate on app startup in production (this repo already disables AutoMigrate unless `ENV=development`).
- Use managed secrets (do NOT put secrets into `docker-compose.yml` or repo). Use Docker secrets if running Swarm/Kubernetes.
- Ensure MySQL only listens on internal network; use Docker network to isolate services.
- Enable TLS between reverse proxy and clients; disable HTTP except for ACME challenge if using Let's Encrypt.
- Use strong `JWT_SECRET` (32+ bytes) and rotate as needed. Consider RS256 with KMS for multi-service setups.
- Limit container privileges and run the app as non-root user inside container.
- Keep dependencies updated and scan images `docker scan`.

Example `docker-compose.prod.yml` (recommended):
- The repo provides a template `docker-compose.dev.yml`. For production, use a separate `docker-compose.prod.yml` with volumes for backups, persistent data, and secure environment injection.

Backup & restore
- Regularly snapshot MySQL data directory or export logical dumps using `mysqldump` to remote storage.
- Rotate logs and backup the S3 content if using S3-compatible storage.

Monitoring
- Add Prometheus exporters and central logging (ELK/Graylog/Loki).
- Add alerting for high error rates, repeated auth failures, and resource exhaustion.

If you want, I can generate a `docker-compose.prod.yml`, systemd unit files, and a sample `deploy.sh` to automate VPS deployment.
