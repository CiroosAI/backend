param(
    [string]$BaseUrl = "http://localhost:8080"
)

# Basic SQL injection payloads to try on search/lookup endpoints.
$payloads = @(
    "' OR '1'='1",
    "' OR 1=1 --",
    "\" OR \"1\"=\"1\"",
    "' UNION SELECT NULL--",
    "'; DROP TABLE users; --"
)

$endpoints = @('/api/register', '/api/login', '/api/products', '/api/user/team')

foreach ($ep in $endpoints) {
    Write-Host "Testing endpoint $ep"
    foreach ($p in $payloads) {
        $body = @{ number = $p; password = "x" } | ConvertTo-Json
        try {
            $r = Invoke-RestMethod -Method Post -Uri "$BaseUrl$ep" -Body $body -ContentType "application/json" -ErrorAction Stop
            Write-Host "Payload $p -> HTTP OK; response: $($r | ConvertTo-Json -Compress)"
        } catch {
            Write-Host "Payload $p -> Error: $($_.Exception.Message)"
        }
    }
}

Write-Host "Expected defense: application uses parameterized queries (GORM) and returns appropriate errors, no DB errors, and no data leakage."