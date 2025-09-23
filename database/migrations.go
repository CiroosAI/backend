package database

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"gorm.io/gorm"
)

// BackupDatabase attempts to create a SQL dump using mysqldump if it's available on PATH.
// It writes to the provided path and returns an error if the command fails.
func BackupDatabase(dsn string, outPath string) error {
	// If mysqldump is not installed, return an informative error
	if _, err := exec.LookPath("mysqldump"); err != nil {
		return fmt.Errorf("mysqldump not found in PATH: %w", err)
	}

	// NOTE: dsn should be parsed by the caller to extract user/host/db or provide a secure wrapper
	// For simplicity we expect caller to supply the appropriate flags in DB_BACKUP_FLAGS env or similar
	args := []string{os.Getenv("DB_BACKUP_FLAGS")}
	// Attempt simple invocation; this can be customized via env
	cmd := exec.CommandContext(context.Background(), "mysqldump", args...)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}
	return nil
}

// RunMigrationsWithBackup runs AutoMigrate after attempting a backup (best-effort).
// It accepts a list of models to migrate. The function attempts a mysqldump backup if
// DB_BACKUP_PATH env is set. It runs migrations inside a transaction where possible.
func RunMigrationsWithBackup(db *gorm.DB, models ...interface{}) error {
	backupPath := os.Getenv("DB_BACKUP_PATH")
	if backupPath != "" {
		// Perform backup asynchronously but wait a short time
		go func() {
			_ = BackupDatabase(os.Getenv("DB_DSN"), backupPath)
		}()
		// allow a small window for the backup to start
		time.Sleep(500 * time.Millisecond)
	}

	// Use AutoMigrate as usual; callers should ensure models are correct and migrations reviewed
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.AutoMigrate(models...); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
