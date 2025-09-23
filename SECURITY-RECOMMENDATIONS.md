# 🔒 Security Recommendations for Production

Dokumen ini berisi rekomendasi keamanan untuk deployment production aplikasi StoneForm.

## 🚨 Critical Security Issues to Address

### 1. Environment Variables & Secrets Management

**❌ Current Issues:**
- Hardcoded secrets di development scripts
- Environment variables tidak terenkripsi
- Tidak ada secret rotation

**✅ Recommendations:**
```bash
# Gunakan Docker secrets atau external secret management
# Contoh dengan Docker secrets:
echo "your_jwt_secret" | docker secret create jwt_secret -
echo "your_db_password" | docker secret create db_password -

# Atau gunakan external secret management seperti:
# - HashiCorp Vault
# - AWS Secrets Manager
# - Azure Key Vault
```

### 2. Database Security

**✅ Implementasi yang sudah baik:**
- Password hashing dengan bcrypt
- Database connection pooling
- TLS support untuk koneksi database

**✅ Additional Recommendations:**
```sql
-- Buat dedicated database users dengan least privilege
CREATE USER 'stoneform_app'@'%' IDENTIFIED BY 'strong_password';
CREATE USER 'stoneform_readonly'@'%' IDENTIFIED BY 'strong_readonly_password';

-- Grant minimal permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON stoneform_db.* TO 'stoneform_app'@'%';
GRANT SELECT ON stoneform_db.* TO 'stoneform_readonly'@'%';

-- Jangan gunakan root user untuk aplikasi
-- Disable root remote access
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
```

### 3. Network Security

**✅ Docker Network Isolation:**
```yaml
# docker-compose.yml sudah menggunakan isolated network
networks:
  app-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

**✅ Additional Recommendations:**
```bash
# Setup firewall rules
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable

# Block direct database access dari internet
sudo ufw deny 3306
sudo ufw deny 6379
```

### 4. Application Security

**✅ Implementasi yang sudah baik:**
- JWT token dengan expiration
- Rate limiting middleware
- Input validation
- CORS configuration
- Security headers middleware

**✅ Additional Recommendations:**

#### A. Password Policy Enhancement
```go
// Tambahkan di password validation
func validatePassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    // Check for complexity
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
    
    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number, and special character")
    }
    
    return nil
}
```

#### B. Account Lockout Enhancement
```go
// Implementasi account lockout yang lebih robust
type LoginAttempt struct {
    UserID    uint      `json:"user_id"`
    IP        string    `json:"ip"`
    Timestamp time.Time `json:"timestamp"`
    Success   bool      `json:"success"`
}

// Rate limiting per IP dan per user
func checkLoginAttempts(userID uint, ip string) error {
    // Check attempts in last 15 minutes
    var attempts int64
    database.DB.Model(&LoginAttempt{}).
        Where("user_id = ? AND ip = ? AND timestamp > ? AND success = false", 
              userID, ip, time.Now().Add(-15*time.Minute)).
        Count(&attempts)
    
    if attempts >= 5 {
        return errors.New("too many failed attempts")
    }
    
    return nil
}
```

### 5. Container Security

**✅ Dockerfile sudah menggunakan:**
- Multi-stage build
- Non-root user
- Distroless base image
- Minimal attack surface

**✅ Additional Recommendations:**
```dockerfile
# Tambahkan security scanning
FROM gcr.io/distroless/static-debian12:nonroot

# Set proper file permissions
RUN chmod 755 /app/server

# Add security labels
LABEL security.scan="true"
LABEL security.level="high"
```

### 6. SSL/TLS Configuration

**✅ Nginx sudah dikonfigurasi dengan:**
- TLS 1.2+ only
- Strong cipher suites
- HSTS headers
- SSL session caching

**✅ Additional Recommendations:**
```nginx
# Tambahkan di nginx.conf
# OCSP stapling
ssl_stapling on;
ssl_stapling_verify on;
ssl_trusted_certificate /etc/nginx/ssl/ca.pem;

# Additional security headers
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
```

### 7. Logging & Monitoring

**✅ Implementasi yang sudah baik:**
- Structured logging
- Request ID tracking
- Error tracking

**✅ Additional Recommendations:**
```go
// Tambahkan security event logging
type SecurityEvent struct {
    EventType   string    `json:"event_type"`
    UserID      uint      `json:"user_id,omitempty"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`
    Timestamp   time.Time `json:"timestamp"`
    Details     string    `json:"details"`
}

func logSecurityEvent(eventType, details string, r *http.Request) {
    event := SecurityEvent{
        EventType: eventType,
        IP:        getClientIP(r),
        UserAgent: r.UserAgent(),
        Timestamp: time.Now(),
        Details:   details,
    }
    
    // Log to structured logger
    log.Printf("SECURITY_EVENT: %+v", event)
    
    // Send to monitoring system (optional)
    // sendToMonitoring(event)
}
```

### 8. Backup & Recovery Security

**✅ Recommendations:**
```bash
# Encrypt database backups
mysqldump -u root -p stoneform_db | gpg --symmetric --cipher-algo AES256 --output backup_$(date +%Y%m%d).sql.gpg

# Store backups in secure location
# - Use encrypted storage
# - Implement access controls
# - Regular backup testing
```

### 9. Dependency Security

**✅ Recommendations:**
```bash
# Regular dependency scanning
go list -json -m all | nancy sleuth

# Update dependencies regularly
go get -u ./...
go mod tidy

# Use tools like:
# - Snyk
# - OWASP Dependency Check
# - GitHub Dependabot
```

### 10. Runtime Security

**✅ Recommendations:**
```bash
# Use security-focused base images
FROM gcr.io/distroless/static-debian12:nonroot

# Implement health checks
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "-health-check"] || exit 1

# Monitor container resources
docker stats --no-stream

# Use container runtime security
# - Enable seccomp
# - Use AppArmor/SELinux
# - Limit container capabilities
```

## 🛡️ Security Checklist

### Pre-Deployment
- [ ] Remove all hardcoded secrets
- [ ] Use strong, unique passwords
- [ ] Enable database TLS
- [ ] Configure firewall rules
- [ ] Setup SSL certificates
- [ ] Enable security headers
- [ ] Configure rate limiting
- [ ] Setup monitoring

### Post-Deployment
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity
- [ ] Regular backup testing
- [ ] Dependency vulnerability scanning
- [ ] Penetration testing
- [ ] Security audit

### Ongoing Maintenance
- [ ] Monthly security updates
- [ ] Quarterly security review
- [ ] Annual penetration testing
- [ ] Regular backup verification
- [ ] Monitor security advisories

## 🚨 Incident Response Plan

### 1. Detection
- Monitor logs for suspicious activity
- Set up alerts for failed login attempts
- Monitor resource usage anomalies

### 2. Response
```bash
# Immediate response
docker compose logs -f app | grep -i "error\|failed\|attack"

# Check for suspicious activity
docker compose exec db mysql -u root -p -e "SELECT * FROM mysql.general_log WHERE command_type='Connect' ORDER BY event_time DESC LIMIT 10;"

# Isolate affected containers
docker compose stop app
```

### 3. Recovery
- Restore from clean backup
- Update all credentials
- Patch vulnerabilities
- Review and strengthen security measures

## 📞 Security Contacts

- **Security Team**: security@yourcompany.com
- **Emergency**: +1-XXX-XXX-XXXX
- **Bug Bounty**: security@yourcompany.com

## 📚 Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Nginx Security Configuration](https://nginx.org/en/docs/http/configuring_https_servers.html)
- [MySQL Security Guidelines](https://dev.mysql.com/doc/refman/8.0/en/security.html)

---

**⚠️ Penting: Dokumen ini harus di-review secara berkala dan disesuaikan dengan kebutuhan spesifik aplikasi dan lingkungan production.**
