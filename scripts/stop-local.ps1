# Stop-local helper: stop and remove dev stack
Write-Host "Stopping Docker Compose dev stack..."
docker compose -f docker-compose.dev.yml down -v
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to stop docker compose stack. Run the command manually to inspect errors."
    exit 1
}
Write-Host "Stopped and removed dev containers and volumes."
