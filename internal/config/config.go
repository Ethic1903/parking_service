package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultConfigFilePath     = "config/local.yaml"
	defaultAppEnv             = "dev"
	defaultHTTPPort           = "8080"
	defaultReadTimeoutSec     = 5
	defaultWriteTimeoutSec    = 10
	defaultIdleTimeoutSec     = 60
	defaultShutdownTimeoutSec = 10
)

type Config struct {
	AppEnv     string
	ConfigFile string
	HTTP       HTTPConfig
}

type HTTPConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	explicitConfigFilePath := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	configFilePath := explicitConfigFilePath
	if configFilePath == "" {
		configFilePath = defaultConfigFilePath
	}

	v.SetConfigFile(configFilePath)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) || os.IsNotExist(err) {
			if explicitConfigFilePath != "" {
				return Config{}, fmt.Errorf("read config file %q: %w", explicitConfigFilePath, err)
			}
		} else {
			return Config{}, fmt.Errorf("read config file %q: %w", configFilePath, err)
		}
	}

	appEnv := strings.TrimSpace(v.GetString("app.env"))
	if appEnv == "" {
		appEnv = defaultAppEnv
	}

	cfg := Config{
		AppEnv: appEnv,
		HTTP: HTTPConfig{
			Port:            normalizePort(v.GetString("http.port")),
			ReadTimeout:     toDurationSec(sanitizePositiveInt(v.GetInt("http.read_timeout_sec"), defaultReadTimeoutSec)),
			WriteTimeout:    toDurationSec(sanitizePositiveInt(v.GetInt("http.write_timeout_sec"), defaultWriteTimeoutSec)),
			IdleTimeout:     toDurationSec(sanitizePositiveInt(v.GetInt("http.idle_timeout_sec"), defaultIdleTimeoutSec)),
			ShutdownTimeout: toDurationSec(sanitizePositiveInt(v.GetInt("http.shutdown_timeout_sec"), defaultShutdownTimeoutSec)),
		},
	}

	if used := v.ConfigFileUsed(); used != "" {
		cfg.ConfigFile = used
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", defaultAppEnv)
	v.SetDefault("http.port", defaultHTTPPort)
	v.SetDefault("http.read_timeout_sec", defaultReadTimeoutSec)
	v.SetDefault("http.write_timeout_sec", defaultWriteTimeoutSec)
	v.SetDefault("http.idle_timeout_sec", defaultIdleTimeoutSec)
	v.SetDefault("http.shutdown_timeout_sec", defaultShutdownTimeoutSec)
}

func normalizePort(raw string) string {
	port := strings.TrimSpace(raw)
	port = strings.TrimPrefix(port, ":")
	if port == "" {
		return defaultHTTPPort
	}

	return port
}

func sanitizePositiveInt(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}

	return value
}

func toDurationSec(value int) time.Duration {
	return time.Duration(value) * time.Second
}
