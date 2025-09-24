# 🚀 RUN VPS - StoneForm Backend & Frontend

Dokumentasi lengkap untuk menjalankan aplikasi StoneForm di VPS Ubuntu/Debian.

## 📋 **PREREQUISITES**

- VPS Ubuntu 20.04+ atau Debian 11+
- Root access atau user dengan sudo privileges
- Domain yang sudah di-point ke VPS (sqcapitall.space)
- Minimal 2GB RAM, 2 CPU cores, 20GB storage

## 🔧 **STEP 1: SETUP VPS**

### 1.1 Update System
```bash
sudo apt update && sudo apt upgrade -y
```

### 1.2 Install Dependencies
```bash
sudo apt install -y curl wget git nginx certbot python3-certbot-nginx
```

### 1.3 Install Docker & Docker Compose
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout dan login ulang
exit
```

## ��️ **STEP 2: SETUP BACKEND**

### 2.1 Clone Repository
```bash
# Clone backend
git clone https://github.com/username/stoneform-backend.git backend
cd backend
```

### 2.2 Setup Environment Variables
```bash
# Copy environment template
cp .env.example .env

# Edit environment variables
nano .env
```

**Isi .env:**
```env
# Database Configuration
DB_HOST=db
DB_PORT=3306
DB_USER=vla_user
DB_PASS=your_secure_password
DB_NAME=vla-sf
DB_ROOT_PASSWORD=your_root_password
DB_TLS=false
DB_TLS_VERIFY=false
DB_PARAMS=charset=utf8mb4&parseTime=True&loc=Local&tls=false&timeout=10s&readTimeout=10s&writeTimeout=10s

# Redis Configuration
REDIS_ADDR=redis:6379
REDIS_PASS=your_redis_password
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# S3 Configuration (Optional)
S3_ENDPOINT=https://s3.amazonaws.com
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=your_bucket_name
S3_REGION=us-east-1

# Pakasir Configuration
PAKASIR_API_KEY=your_pakasir_api_key
PAKASIR_PROJECT=your_project_name
PAKASIR_BASE_URL=https://app.pakasir.com

# Klikpay Configuration
KLIKPAY_API_KEY=your_klikpay_api_key
KLIKPAY_PROJECT=your_project_name
KLIKPAY_BASE_URL=https://app.klikpay.com

# Security Configuration
CORS_ORIGINS=https://sqcapitall.space,https://www.sqcapitall.space
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=3600

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json

# Cron Configuration
CRON_KEY=your_cron_secret_key

# Application Configuration
ENV=production
PORT=8080
APP_PORT=8080
```

### 2.3 Start Backend Services
```bash
# Start backend dengan Docker Compose
docker compose up -d

# Cek status
docker compose ps

# Cek logs
docker compose logs -f
```

### 2.4 Setup Database
```bash
# Cek database connection
docker exec vla-mysql mysql -u root -p$DB_ROOT_PASSWORD -e "SHOW DATABASES;"

# Run migrations (otomatis via Docker)
# Cek apakah tables sudah terbuat
docker exec vla-mysql mysql -u root -p$DB_ROOT_PASSWORD -e "USE \`vla-sf\`; SHOW TABLES;"
```

## 🎨 **STEP 3: SETUP FRONTEND**

### 3.1 Clone Frontend Repository
```bash
# Clone frontend
git clone https://github.com/username/stoneform-frontend.git frontend
cd frontend
```

### 3.2 Install Dependencies
```bash
# Install Node.js (jika belum ada)
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install dependencies
npm install
```

### 3.3 Setup Environment Variables
```bash
# Edit .env.local
nano .env.local
```

**Isi .env.local:**
```env
NEXT_PUBLIC_API_URL=https://sqcapitall.space
NEXT_PUBLIC_APP_NAME=StoneForm
NEXT_PUBLIC_APP_VERSION=1.0.0
```

### 3.4 Update Next.js Config
```bash
# Edit next.config.js
nano next.config.js
```

**Isi next.config.js:**
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  outputFileTracingRoot: __dirname,
  images: {
    unoptimized: true
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
}

module.exports = nextConfig
```

### 3.5 Start Frontend dengan PM2
```bash
# Install PM2
sudo npm install -g pm2

# Start frontend
sudo pm2 start "npm start" --name "frontend" --cwd /home/ubuntu/frontend

# Save PM2 config
sudo pm2 save

# Setup PM2 untuk auto-start
sudo pm2 startup

# Cek status
sudo pm2 status
```

## �� **STEP 4: SETUP NGINX**

### 4.1 Create Nginx Config
```bash
# Create nginx config
sudo nano /etc/nginx/sites-available/sqcapitall.space
```

**Isi nginx config:**
```nginx
server {
    listen 80;
    server_name sqcapitall.space www.sqcapitall.space;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sqcapitall.space www.sqcapitall.space;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/sqcapitall.space/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sqcapitall.space/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # API routes (backend)
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Admin routes
    location /admin {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Frontend routes
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### 4.2 Enable Site
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/sqcapitall.space /etc/nginx/sites-enabled/

# Test nginx config
sudo nginx -t

# Start nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

## 🔒 **STEP 5: SETUP SSL CERTIFICATE**

### 5.1 Generate SSL Certificate
```bash
# Generate SSL certificate
sudo certbot --nginx -d sqcapitall.space -d www.sqcapitall.space
```

### 5.2 Test SSL
```bash
# Test SSL
curl https://sqcapitall.space
curl https://sqcapitall.space/api/health
```

## ✅ **STEP 6: VERIFICATION**

### 6.1 Test Backend
```bash
# Test API endpoints
curl https://sqcapitall.space/api/health
curl https://sqcapitall.space/api/payment_info

# Test dengan VLA key
curl -H "X-VLA-KEY: VLA010124" https://sqcapitall.space/api/payment_info
```

### 6.2 Test Frontend
```bash
# Test frontend
curl https://sqcapitall.space
curl https://sqcapitall.space/admin
curl https://sqcapitall.space/admin/login
```

### 6.3 Test di Browser
- Buka https://sqcapitall.space
- Buka https://sqcapitall.space/admin
- Buka https://sqcapitall.space/admin/login

## �� **STEP 7: MAINTENANCE**

### 7.1 Cek Status Services
```bash
# Cek backend
docker compose ps

# Cek frontend
sudo pm2 status

# Cek nginx
sudo systemctl status nginx
```

### 7.2 Cek Logs
```bash
# Backend logs
docker compose logs -f

# Frontend logs
sudo pm2 logs frontend

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### 7.3 Restart Services
```bash
# Restart backend
docker compose restart

# Restart frontend
sudo pm2 restart frontend

# Restart nginx
sudo systemctl restart nginx
```

### 7.4 Update Application
```bash
# Update backend
cd backend
git pull
docker compose up -d --build

# Update frontend
cd frontend
git pull
sudo pm2 restart frontend
```

## 🚨 **TROUBLESHOOTING**

### Common Issues:

1. **Port 80/443 already in use**
   ```bash
   sudo lsof -i :80
   sudo lsof -i :443
   sudo systemctl stop apache2  # jika ada
   ```

2. **Database connection failed**
   ```bash
   docker compose logs vla-mysql
   docker compose restart db
   ```

3. **Frontend not loading**
   ```bash
   sudo pm2 logs frontend
   sudo pm2 restart frontend
   ```

4. **SSL certificate issues**
   ```bash
   sudo certbot certificates
   sudo certbot renew --dry-run
   ```

## �� **MONITORING**

### 7.1 Setup Monitoring
```bash
# Install htop untuk monitoring
sudo apt install htop

# Monitor resources
htop
```

### 7.2 Setup Log Rotation
```bash
# Setup logrotate untuk nginx
sudo nano /etc/logrotate.d/nginx
```

## 🔐 **SECURITY**

### 8.1 Firewall Setup
```bash
# Setup UFW firewall
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

### 8.2 Regular Updates
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Update Docker images
docker compose pull
docker compose up -d
```

## 📝 **NOTES**

- Pastikan domain sudah di-point ke VPS sebelum setup SSL
- Backup database secara berkala
- Monitor disk space dan memory usage
- Update dependencies secara berkala
- Test aplikasi setelah setiap update

## 🆘 **SUPPORT**

Jika ada masalah, cek:
1. Logs aplikasi
2. Status services
3. Network connectivity
4. SSL certificate validity
5. Database connection

---

**Happy Deploying! 🚀**