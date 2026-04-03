package storage

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultDBDriver         = "sqlite"
	defaultPostgresHost     = "localhost"
	defaultPostgresPort     = "5432"
	defaultPostgresDBName   = "parking"
	defaultPostgresUser     = "parking"
	defaultPostgresPassword = "parking"
	defaultPostgresSSLMode  = "disable"
	defaultSQLitePath       = "parking.db"
)

type Config struct {
	Driver   string
	Postgres PostgresConfig
	SQLite   SQLiteConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	SSLMode  string
}

type SQLiteConfig struct {
	Path string
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("db.driver", defaultDBDriver)
	v.SetDefault("db.postgres.host", defaultPostgresHost)
	v.SetDefault("db.postgres.port", defaultPostgresPort)
	v.SetDefault("db.postgres.dbname", defaultPostgresDBName)
	v.SetDefault("db.postgres.user", defaultPostgresUser)
	v.SetDefault("db.postgres.password", defaultPostgresPassword)
	v.SetDefault("db.postgres.sslmode", defaultPostgresSSLMode)
	v.SetDefault("db.sqlite.path", defaultSQLitePath)
}

func LoadConfig(v *viper.Viper) Config {
	return Config{
		Driver: strings.ToLower(strings.TrimSpace(v.GetString("db.driver"))),
		Postgres: PostgresConfig{
			Host:     strings.TrimSpace(v.GetString("db.postgres.host")),
			Port:     strings.TrimSpace(v.GetString("db.postgres.port")),
			DBName:   strings.TrimSpace(v.GetString("db.postgres.dbname")),
			User:     strings.TrimSpace(v.GetString("db.postgres.user")),
			Password: v.GetString("db.postgres.password"),
			SSLMode:  strings.TrimSpace(v.GetString("db.postgres.sslmode")),
		},
		SQLite: SQLiteConfig{
			Path: strings.TrimSpace(v.GetString("db.sqlite.path")),
		},
	}
}
