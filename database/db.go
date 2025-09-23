package database

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect connects to the database with secure defaults, pooling and retry.
// It supports role-based connections by using DB_ROLE env var ("read" or "write").
func Connect() (*gorm.DB, error) {
	if DB != nil {
		return DB, nil
	}

	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	user := getenv("DB_USER", "root")
	pass := getenv("DB_PASS", "")
	name := getenv("DB_NAME", "v1")
	params := getenv("DB_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local")

	// Allow explicit DSN override
	dsn := os.Getenv("DB_DSN")

	// Allow role override: "read" will try DB_READ_USER/DB_READ_PASS, "write" uses DB_USER
	role := strings.ToLower(getenv("DB_ROLE", "write"))
	if role == "read" {
		ruser := getenv("DB_READ_USER", "")
		rpass := getenv("DB_READ_PASS", "")
		if ruser != "" {
			user = ruser
			pass = rpass
		}
	}

	if dsn == "" {
		// Ensure TLS/timeout params are present to enforce encrypted connections and timeouts
		// Add defaults for timeouts and parseTime if not present
		if !strings.Contains(params, "tls=") {
			// Accept TLS mode via env DB_TLS (skip, preferred, true)
			tlsMode := getenv("DB_TLS", "true")
			if tlsMode == "true" || tlsMode == "preferred" {
				// use tls=true (requires server to support TLS). For strict verification, user can set DB_TLS=verify
				if getenv("DB_TLS_VERIFY", "false") == "true" {
					// we'll register a custom TLS config below and reference it by name
					params = params + "&tls=custom"
				} else {
					params = params + "&tls=true"
				}
			}
		}
		// connection timeouts
		if !strings.Contains(params, "timeout=") {
			params = params + "&timeout=10s"
		}
		if !strings.Contains(params, "readTimeout=") {
			params = params + "&readTimeout=10s"
		}
		if !strings.Contains(params, "writeTimeout=") {
			params = params + "&writeTimeout=10s"
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, pass, host, port, name, params)
	}

	// Debugging: log the DSN (without password) to help troubleshoot connection issues
	safeDSN := dsn
	if pass != "" {
		safeDSN = strings.Replace(safeDSN, pass, "******", 1)
	}
	log.Printf("[database] using DSN: %s", safeDSN)

	// Optionally register a custom TLS config named "custom" for strict certificate validation
	if strings.Contains(dsn, "tls=custom") {
		// Load CA bundle path from env
		caPath := getenv("DB_TLS_CA_PATH", "")
		tlsCfg := &tls.Config{}
		if caPath != "" {
			caCert, err := ioutil.ReadFile(caPath)
			if err != nil {
				return nil, fmt.Errorf("failed reading DB TLS CA file: %w", err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(caCert) {
				return nil, errors.New("failed to append CA certs")
			}
			tlsCfg.RootCAs = pool
		}
		// Optionally load client cert/key
		clientCert := getenv("DB_TLS_CLIENT_CERT", "")
		clientKey := getenv("DB_TLS_CLIENT_KEY", "")
		if clientCert != "" && clientKey != "" {
			cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client cert/key: %w", err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}

		// Register with go-sql-driver/mysql driver
		mysqldriver.RegisterTLSConfig("custom", tlsCfg)
	}

	// GORM logger: verbose in development
	var gormLogger logger.Interface
	if strings.ToLower(getenv("ENV", "development")) == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Retry connection with exponential backoff
	maxRetries := atoi(getenv("DB_CONNECT_RETRIES", "5"))
	var db *gorm.DB
	var err error
	backoff := time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		db, err = gorm.Open(gormmysql.Open(dsn), &gorm.Config{Logger: gormLogger})
		if err == nil {
			break
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	if err != nil {
		return nil, err
	}

	// Configure connection pool on the underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxOpen := atoi(getenv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle := atoi(getenv("DB_MAX_IDLE_CONNS", "25"))
	maxLifetimeSec := atoi(getenv("DB_CONN_MAX_LIFETIME", "3600"))

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetimeSec) * time.Second)

	// Optional connection validation
	if getenv("DB_PING_ON_CONNECT", "true") == "true" {
		if err := pingWithTimeout(sqlDB, 5*time.Second); err != nil {
			return nil, fmt.Errorf("database ping failed: %w", err)
		}
	}

	DB = db
	return DB, nil
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	if v <= 0 {
		return 0
	}
	return v
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	type pinger interface {
		Ping() error
	}
	// Use a goroutine with timeout to avoid blocking
	ch := make(chan error, 1)
	go func() {
		ch <- db.Ping()
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("ping timeout after %s", timeout)
	}
}
