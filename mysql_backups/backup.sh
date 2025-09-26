#!/bin/bash

# Konfigurasi
CONTAINER="sf-mysql-dev"
USER="root"
PASSWORD="rootpassword"
BACKUP_DIR="$HOME/mysql_backups"
DATE=$(date +"%Y-%m-%d_%H-%M-%S")
FILENAME="$BACKUP_DIR/backup-$DATE.sql"

# Jalankan backup
mkdir -p $BACKUP_DIR
docker exec $CONTAINER mysqldump -u $USER -p$PASSWORD --all-databases > $FILENAME

echo "✅ Backup selesai: $FILENAME"

