package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrationFS embed.FS

func Run(ctx context.Context, db *sql.DB, driver string) error {
	dialect := gooseDialect(driver)
	if dialect == "" {
		return fmt.Errorf("unsupported db driver %q", driver)
	}

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}
	goose.SetBaseFS(migrationFS)

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return err
	}

	return nil
}

func gooseDialect(driver string) string {
	switch driver {
	case "postgres":
		return "postgres"
	case "sqlite":
		return "sqlite3"
	default:
		return ""
	}
}
