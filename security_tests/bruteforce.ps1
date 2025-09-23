param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Number = "8123456789",
    [string]$Wordlist = "./security_tests/wordlist.txt",
    [int]$DelayMs = 500
)

Write-Host "Brute force test against $BaseUrl/api/login"
$attempt = 0
Get-Content $Wordlist | ForEach-Object {
    $attempt++
    $pwd = $_.Trim()
    if ($pwd -eq "") { return }
    $body = @{ number = $Number; password = $pwd } | ConvertTo-Json
    $resp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/login" -Body $body -ContentType "application/json" -ErrorAction SilentlyContinue
    Write-Host "[$attempt] Tried password: $pwd -> HTTP: $($LASTEXITCODE)" -NoNewline
    if ($resp -and $resp.success -eq $true) {
        Write-Host " => SUCCESS! found password: $pwd"
        return
    }
    else {
        Write-Host ""
    }
    Start-Sleep -Milliseconds $DelayMs
}
Write-Host "Brute force test completed. Expected defenses: rate-limit / account lockout / consistent error messages. Check server logs for lockouts."