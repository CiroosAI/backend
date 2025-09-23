# FOR LOCAL DEVELOPMENT ONLY — DO NOT USE IN PRODUCTION
# Run-local helper: starts docker dev stack, waits for DB/Redis, then runs the app on host
# Usage: Open PowerShell in project root and run: .\scripts\run-local.ps1

param(
    [int]$DbHostPort = 3307,
    [int]$RedisPort = 6379,
    [int]$WaitTimeoutSec = 120
)

function Wait-Port($HostName, $Port, $timeoutSec) {
    $start = Get-Date
    while ($true) {
        $elapsed = (Get-Date) - $start
        if ($elapsed.TotalSeconds -ge $timeoutSec) {
            return $false
        }
        $res = Test-NetConnection -ComputerName $HostName -Port $Port -WarningAction SilentlyContinue
        if ($res.TcpTestSucceeded) {
            return $true
        }
        Start-Sleep -Seconds 1
    }
}

Write-Host "Starting Docker Compose dev stack..."
docker compose -f docker-compose.dev.yml up -d --build
if ($LASTEXITCODE -ne 0) {
    Write-Error "docker compose failed. Check Docker Desktop and try again."
    exit 1
}

Write-Host "Waiting for MySQL on 127.0.0.1:$DbHostPort..."
if (-not (Wait-Port -host '127.0.0.1' -port $DbHostPort -timeoutSec $WaitTimeoutSec)) {
    Write-Error "Timed out waiting for MySQL on 127.0.0.1:$DbHostPort. Check container logs: docker compose -f docker-compose.dev.yml logs db"
    exit 1
}
Write-Host "MySQL is listening on 127.0.0.1:$DbHostPort"

Write-Host "Waiting for Redis on 127.0.0.1:$RedisPort..."
if (-not (Wait-Port -host '127.0.0.1' -port $RedisPort -timeoutSec $WaitTimeoutSec)) {
    Write-Error "Timed out waiting for Redis on 127.0.0.1:$RedisPort. Check container logs: docker compose -f docker-compose.dev.yml logs redis"
    exit 1
}
Write-Host "Redis is listening on 127.0.0.1:$RedisPort"

Write-Host "Starting the Go app locally with development env..."
$env:ENV = "development"
$env:DB_HOST = "127.0.0.1"
$env:DB_PORT = "$DbHostPort"
$env:DB_USER = "root"
$env:DB_PASS = "rootpassword"
$env:DB_NAME = "sf"
$env:JWT_SECRET = "supersecretjwtkey"
$env:REDIS_ADDR = "127.0.0.1:$RedisPort"
$env:S3_ENDPOINT=https://i8h4.sg.idrivee2-24.com
$env:S3_ACCESS_KEY=xiiZL72Ie1vnxyb63dcH
$env:S3_SECRET_KEY=EUMhvuNqSOvgXQlMNuQrjQ3CktyrKeppu4of3DEQ

Write-Host "Environment ready. Running: go run main.go (child cmd with env)"

# Run Go in a child cmd.exe process with env vars set for that process only. This avoids godotenv overwriting host env.
$cmd = "set DB_HOST=127.0.0.1 && set DB_PORT=$DbHostPort && set DB_USER=root && set DB_PASS=rootpassword && set DB_NAME=sf && set JWT_SECRET=supersecretjwtkey && set REDIS_ADDR=127.0.0.1:$RedisPort && go run main.go"
cmd.exe /c $cmd
exit $LASTEXITCODE
