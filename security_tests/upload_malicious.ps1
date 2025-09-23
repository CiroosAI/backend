param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIiLCJleHAiOjE3NTgyMTkzNjksImlhdCI6MTc1ODIxODQ2OSwiaWQiOjcsImlzcyI6IiIsImp0aSI6IjRjMzg5M2RlMThlODY4ZWU4ZTBjYmZiNjA4ZDU3YjY3IiwibmJmIjoxNzU4MjE4NDY5LCJyb2xlIjoidXNlciJ9.XCLJHKvl7t12QPG0U1SthsreWKLvh2fukkvTT9VTsVg"
)

# Create a fake JS file disguised as image
$malicious = "<script>alert('xss')</script>"
$bytes = [System.Text.Encoding]::UTF8.GetBytes($malicious)
$filePath = Join-Path -Path $PWD -ChildPath "security_tests/malicious.jpg"
[IO.File]::WriteAllBytes($filePath, $bytes)

$headers = @{}
if ($AuthToken -ne "") { $headers.Add("Authorization", "Bearer $AuthToken") }

Write-Host "Uploading malicious.jpg to /api/users/forum/submit (multipart/form-data)"
try {
    $resp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/users/forum/submit" -InFile $filePath -ContentType "multipart/form-data" -Headers $headers -ErrorAction Stop
    Write-Host "Response: $($resp | ConvertTo-Json -Compress)"
} catch {
    Write-Host "Upload error: $($_.Exception.Message)"
}

Write-Host "Expected defenses: MIME sniffing reject, re-encoding to valid image, size limits. If uploaded, then review server-side upload validation and re-encoding."