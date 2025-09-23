# 🚀 Production Deployment Guide

Panduan lengkap untuk deploy aplikasi vla ke VPS Ubuntu/Debian menggunakan Docker Compose.

## 📋 Prerequisites

- VPS dengan Ubuntu 20.04+ atau Debian 11+
- Minimal 2GB RAM, 2 CPU cores, 20GB storage
- Root access atau user dengan sudo privileges
- Domain name (opsional, untuk HTTPS)

## 🔧 Step 1: Setup Server

### Update sistem dan install dependencies

```bash
# Update package list
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y curl wget git unzip software-properties-common apt-transport-https ca-certificates gnupg lsb-release

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

## 📁 Step 2: Clone Repository

```bash
# Clone repository
git clone <your-repository-url>
cd vla-backend

# Atau jika menggunakan SSH
git clone git@github.com:yourusername/vla-backend.git
cd vla-backend
```

## ⚙️ Step 3: Environment Configuration

### Buat file .env

```bash
# Copy template environment file
cp env.example .env

# Edit dengan editor favorit
nano .env
```

### Konfigurasi .env untuk production:

```bash
# Application Environment
ENV=production

# Server Configuration
PORT=8080
APP_PORT=8080

# Database Configuration
DB_HOST=db
DB_PORT=3306
DB_USER=vla_user
DB_PASS=your_very_secure_db_password_here
DB_NAME=vla_db
DB_ROOT_PASSWORD=your_very_secure_root_password_here

# JWT Configuration
JWT_SECRET=your_very_secure_jwt_secret_key_minimum_32_characters_long

# Redis Configuration
REDIS_ADDR=redis:6379
REDIS_PASS=your_very_secure_redis_password_here
REDIS_DB=0

# S3 Configuration (jika menggunakan file upload)
S3_ENDPOINT=https://your-s3-endpoint.com
S3_ACCESS_KEY=your_s3_access_key
S3_SECRET_KEY=your_s3_secret_key
S3_BUCKET=your_bucket_name
S3_REGION=your_region
```

### Generate secure passwords:

```bash
# Generate random passwords
openssl rand -base64 32  # Untuk JWT_SECRET
openssl rand -base64 32  # Untuk DB_PASS
openssl rand -base64 32  # Untuk DB_ROOT_PASSWORD
openssl rand -base64 32  # Untuk REDIS_PASS
```

## 🐳 Step 4: Deploy dengan Docker Compose

### Deploy aplikasi tanpa Nginx (langsung expose port)

```bash
# Build dan start services
docker compose up -d --build

# Check status
docker compose ps

# View logs
docker compose logs -f
```

### Deploy dengan Nginx reverse proxy (recommended)

```bash
# Deploy dengan Nginx
docker compose --profile nginx up -d --build

# Check status
docker compose ps
```

## 🔒 Step 5: Setup SSL/HTTPS (Opsional)

### Menggunakan Let's Encrypt dengan Certbot

```bash
# Install certbot
sudo apt install -y certbot

# Stop nginx sementara
docker compose stop nginx

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Copy certificates ke project
sudo mkdir -p ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ssl/key.pem
sudo chown -R $USER:$USER ssl/

# Restart nginx
docker compose start nginx
```

### Setup auto-renewal

```bash
# Edit crontab
sudo crontab -e

# Tambahkan baris berikut untuk auto-renewal
0 12 * * * /usr/bin/certbot renew --quiet && docker compose restart nginx
```

## 🔄 Step 6: Setup Auto-Start

### Menggunakan Docker Compose restart policy

File `docker-compose.yml` sudah dikonfigurasi dengan `restart: unless-stopped`, jadi aplikasi akan otomatis restart jika server reboot.

### Menggunakan systemd (opsional)

```bash
# Buat systemd service
sudo nano /etc/systemd/system/vla.service
```

```ini
[Unit]
Description=vla Application
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/path/to/your/vla-backend
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
# Enable dan start service
sudo systemctl enable vla.service
sudo systemctl start vla.service
```

## 📊 Step 7: Monitoring dan Maintenance

### Check aplikasi status

```bash
# Check container status
docker compose ps

# Check logs
docker compose logs -f app
docker compose logs -f db
docker compose logs -f redis

# Check resource usage
docker stats
```

### Health check

```bash
# Test aplikasi
curl http://localhost:8080/health

# Test dengan domain (jika sudah setup)
curl https://yourdomain.com/health
```

### Backup database

```bash
# Buat script backup
nano backup-db.sh
```

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup"
mkdir -p $BACKUP_DIR

docker compose exec -T db mysqldump -u root -p$DB_ROOT_PASSWORD $DB_NAME > $BACKUP_DIR/vla_$DATE.sql
gzip $BACKUP_DIR/vla_$DATE.sql

# Hapus backup lebih dari 7 hari
find $BACKUP_DIR -name "vla_*.sql.gz" -mtime +7 -delete
```

```bash
# Buat executable
chmod +x backup-db.sh

# Setup cron job untuk backup harian
crontab -e
# Tambahkan: 0 2 * * * /path/to/backup-db.sh
```

## 🔄 Step 8: Update Aplikasi

### Zero-downtime update

```bash
# Pull perubahan terbaru
git pull origin main

# Rebuild dan restart dengan rolling update
docker compose up -d --build --force-recreate

# Atau untuk update yang lebih smooth:
docker compose build app
docker compose up -d app
```

### Rollback jika ada masalah

```bash
# Rollback ke commit sebelumnya
git log --oneline
git checkout <previous-commit-hash>
docker compose up -d --build
```

## 🛠️ Troubleshooting

### Common issues dan solusi

#### 1. Database connection error

```bash
# Check database logs
docker compose logs db

# Check database status
docker compose exec db mysql -u root -p -e "SHOW DATABASES;"
```

#### 2. Application tidak start

```bash
# Check application logs
docker compose logs app

# Check environment variables
docker compose exec app env | grep DB_
```

#### 3. Port sudah digunakan

```bash
# Check port usage
sudo netstat -tlnp | grep :8080

# Kill process yang menggunakan port
sudo kill -9 <PID>
```

#### 4. Out of memory

```bash
# Check memory usage
free -h
docker stats

# Restart services
docker compose restart
```

### Log locations

- Application logs: `docker compose logs app`
- Database logs: `docker compose logs db`
- Redis logs: `docker compose logs redis`
- Nginx logs: `docker compose logs nginx`

## 🔐 Security Best Practices

### 1. Firewall setup

```bash
# Install ufw
sudo apt install -y ufw

# Configure firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

### 2. Database security

- Gunakan password yang kuat
- Jangan expose database port ke public
- Regular backup
- Monitor connection logs

### 3. Application security

- Update dependencies secara regular
- Monitor logs untuk suspicious activity
- Gunakan HTTPS
- Implement rate limiting

## 📈 Performance Optimization

### 1. Database optimization

```bash
# Check database performance
docker compose exec db mysql -u root -p -e "SHOW PROCESSLIST;"

# Optimize MySQL configuration
docker compose exec db mysql -u root -p -e "SET GLOBAL innodb_buffer_pool_size = 256M;"
```

### 2. Redis optimization

```bash
# Check Redis memory usage
docker compose exec redis redis-cli info memory

# Monitor Redis performance
docker compose exec redis redis-cli monitor
```

### 3. Application scaling

```bash
# Scale application instances
docker compose up -d --scale app=3

# Load balancer configuration di nginx.conf
upstream backend {
    server app:8080;
    server app:8080;
    server app:8080;
}
```

## 📞 Support

Jika mengalami masalah:

1. Check logs: `docker compose logs -f`
2. Check status: `docker compose ps`
3. Restart services: `docker compose restart`
4. Check resource usage: `docker stats`

## 🎯 Quick Commands Reference

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# Restart services
docker compose restart

# View logs
docker compose logs -f

# Check status
docker compose ps

# Update application
git pull && docker compose up -d --build

# Backup database
docker compose exec db mysqldump -u root -p$DB_ROOT_PASSWORD $DB_NAME > backup.sql

# Access database
docker compose exec db mysql -u root -p

# Access Redis
docker compose exec redis redis-cli

# Check health
curl http://localhost:8080/health
```

---

**Selamat! Aplikasi vla sudah berjalan di production! 🎉**
