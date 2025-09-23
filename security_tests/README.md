Security test scripts for StoneForm backend

Overview
- These scripts are defensive security tests you can run locally against your running server to simulate common attacks (authentication brute force, JWT replay, SQL injection attempts, malicious uploads, rate-limit bypass and timing tests).
- They are intended for testing your own systems only.

How to run
- Ensure your backend is running locally (e.g., `go run main.go`) and note `BASE_URL` (default `http://localhost:8080`).
- Open PowerShell in the project root and run scripts like:

  # Example
  $base = 'http://localhost:8080'
  .\security_tests\bruteforce.ps1 -BaseUrl $base -Number '0812...' -Wordlist .\security_tests\wordlist.txt

Files
- `bruteforce.ps1` - brute force login attempts using a wordlist.
- `jwt_replay.ps1` - logs in, then uses access token after logout to test revocation.
- `sql_injection.ps1` - sends common SQL payloads to suspect endpoints.
- `upload_malicious.ps1` - attempts to upload a disguised malicious file to forum endpoint.
- `wordlist.txt` - sample small wordlist (you should replace with your own for real tests).

Notes
- These scripts are not aggressive DDoS tools; they run sequentially and include delays. Adjust parameters carefully.
- Review and update endpoint paths and JSON structures if your routes differ.
- Expected defenses depend on your configuration. The README inside each script includes expected responses and remediation suggestions.

CI / Automation
- You can wrap these scripts in a CI job (GitHub Actions) that's gated for staging environments only.
- Convert to Go integration tests using `net/http/httptest` if you want to exercise handlers without a live server.

Safety
- Only run these against systems you own or have explicit permission to test.

Run:
.\security_tests\run_all.ps1 -BaseUrl 'http://localhost:8080' -Number '8123456789' -Password '123456'
