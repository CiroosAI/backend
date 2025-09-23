param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Number = "8123456789",
    [string]$Password = "password"
)

# Log in and get tokens
$loginBody = @{ number = $Number; password = $Password } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/login" -Body $loginBody -ContentType "application/json" -ErrorAction SilentlyContinue
if (-not $login -or -not $login.success) {
    Write-Host "Login failed; cannot test replay. Ensure valid credentials or adjust test values."
    return
}
$access = $login.data.access_token
$refresh = $login.data.refresh_token
Write-Host "Got access token (len $($access.Length)) and refresh token"

# Logout using the refresh token (if endpoint exists)
$logoutBody = @{ refresh_token = $refresh } | ConvertTo-Json
$logout = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/logout" -Body $logoutBody -ContentType "application/json" -ErrorAction SilentlyContinue
Write-Host "Logout response: $($logout | ConvertTo-Json -Compress)"

# Attempt to use old access token after logout
$headers = @{ Authorization = "Bearer $access" }
$info = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/users/info" -Headers $headers -ErrorAction SilentlyContinue
Write-Host "Access attempt after logout -> response: $($info | ConvertTo-Json -Compress)"
Write-Host "Expected: after logout/revoke, access token should be rejected (401). If it succeeds, implement access-token revocation (Redis DB) and check ValidateAccessToken."