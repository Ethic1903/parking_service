package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func OpenDB(cfg Config) (*sql.DB, string, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	switch driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.DBName,
			cfg.Postgres.User,
			cfg.Postgres.Password,
			cfg.Postgres.SSLMode,
		)
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, "", err
		}
		return db, driver, nil
	case "sqlite":
		db, err := sql.Open("sqlite", cfg.SQLite.Path)
		if err != nil {
			return nil, "", err
		}
		return db, driver, nil
	default:
		return nil, "", fmt.Errorf("unsupported DB_DRIVER %q", cfg.Driver)
	}
}
