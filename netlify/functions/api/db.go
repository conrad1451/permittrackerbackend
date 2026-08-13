// --- netlify/functions/api/db.go ---

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	// The `pq` package is a pure Go PostgreSQL driver for `database/sql`.
	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// getDB returns a lazily-initialized, connection-pooled *sql.DB.
// Netlify Functions reuse warm containers between invocations, so we
// keep the pool as a package-level singleton instead of opening a new
// connection on every request. Keep MaxOpenConns low.
//
// Every handler MUST call getDB() to obtain the connection instead of
// reading the package-level `db` variable directly - that var is only
// populated the first time getDB() runs.
func getDB() (*sql.DB, error) {
	var initErr error
	dbOnce.Do(func() {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")

		if port == "" {
			port = "5432"
		}

		// sslmode=require is the safe default for hosted Postgres providers
		// (Supabase, Neon, RDS, etc). Set DB_SSLMODE=disable for local dev.
		sslmode := os.Getenv("DB_SSLMODE")
		if sslmode == "" {
			sslmode = "require"
		}

		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
			host, port, user, pass, name, sslmode,
		)

		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			initErr = err
			return
		}

		// Serverless-friendly pool settings: small pool, short idle time,
		// so we don't hold connections open across cold starts / exceed
		// the DB's max connection limit under concurrent invocations.
		conn.SetMaxOpenConns(5)
		conn.SetMaxIdleConns(2)
		conn.SetConnMaxLifetime(3 * time.Minute)
		conn.SetConnMaxIdleTime(1 * time.Minute)

		if err := conn.Ping(); err != nil {
			initErr = fmt.Errorf("db ping failed: %w", err)
			return
		}

		log.Println("database connection pool initialized")
		db = conn
	})

	if initErr != nil {
		return nil, initErr
	}
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db, nil
}