param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Number = "8123456789",
    [string]$Password = "password"
)

Write-Host "Running basic security checks against $BaseUrl"
Write-Host "1) Brute force (first 5 words)"
.\security_tests\bruteforce.ps1 -BaseUrl $BaseUrl -Number $Number -Wordlist .\security_tests\wordlist.txt -DelayMs 200

Write-Host "2) JWT replay"
.\security_tests\jwt_replay.ps1 -BaseUrl $BaseUrl -Number $Number -Password $Password

Write-Host "3) SQL injection probes"
.\security_tests\sql_injection.ps1 -BaseUrl $BaseUrl

Write-Host "4) Upload malicious payload"
# Attempt without auth first
.\security_tests\upload_malicious.ps1 -BaseUrl $BaseUrl -AuthToken ""

Write-Host "Test run complete. Review outputs for any unexpected success responses."