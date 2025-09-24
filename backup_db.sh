#!/bin/bash

# Konfigurasi
DB_NAME="vla-sf"
DB_USER="root"
DB_PASS="vlaroot123"
BACKUP_DIR="/home/ubuntu/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Buat direktori backup
mkdir -p $BACKUP_DIR

# Backup database
docker exec vla-mysql mysqldump -u $DB_USER -p$DB_PASS $DB_NAME | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Hapus backup lama (lebih dari 7 hari)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: backup_$DATE.sql.gz"