package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpServerPort     string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	DefaultListLimit   int
	MaxListLimit       int
	LogLevel           string
	HttpReadTimeout    time.Duration
	HttpWriteTimeout   time.Duration
	HttpIdleTimeout    time.Duration

	// Database pool settings
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnIdleTime time.Duration
	DBMaxConnLifetime time.Duration
}

func LoadConfig(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		// .env file is optional, continue with environment variables
	}

	config := &Config{
		HttpServerPort: getEnv("HTTP_SERVER_PORT", "8080"),
		DatabaseURL:    BuildDatabaseURL(),

		JWTSecret:          getEnv("JWT_SECRET", ""),
		AccessTokenExpiry:  time.Duration(getEnvAsInt("ACCESS_TOKEN_EXPIRY_MINUTES", 15)) * time.Minute,
		RefreshTokenExpiry: time.Duration(getEnvAsInt("REFRESH_TOKEN_EXPIRY_HOURS", 168)) * time.Hour,
		DefaultListLimit:   getEnvAsInt("DEFAULT_LIST_LIMIT", 20),
		MaxListLimit:       getEnvAsInt("MAX_LIST_LIMIT", 100),
		LogLevel:           getEnv("LOG_LEVEL", "info"),

		HttpReadTimeout:  time.Duration(getEnvAsInt("HTTP_READ_TIMEOUT_SECONDS", 15)) * time.Second,
		HttpWriteTimeout: time.Duration(getEnvAsInt("HTTP_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
		HttpIdleTimeout:  time.Duration(getEnvAsInt("HTTP_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,

		DBMaxConns:        int32(getEnvAsInt("DB_MAX_CONNS", 10)),
		DBMinConns:        int32(getEnvAsInt("DB_MIN_CONNS", 2)),
		DBMaxConnIdleTime: time.Duration(getEnvAsInt("DB_MAX_CONN_IDLE_TIME_MINUTES", 30)) * time.Minute,
		DBMaxConnLifetime: time.Duration(getEnvAsInt("DB_MAX_CONN_LIFETIME_MINUTES", 60)) * time.Minute,
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// BuildDatabaseURL constructs a PostgreSQL connection URL from individual
// environment variables (POSTGRES_HOST, POSTGRES_PORT, etc.) if DATABASE_URL
// is not set. This logic is shared between the server and the migrate command.
func BuildDatabaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		PostgresHost := getEnv("POSTGRES_HOST", "localhost")
		PostgresPort := getEnv("POSTGRES_PORT", "5432")
		PostgresUser := getEnv("POSTGRES_USER", "user")
		PostgresPassword := getEnv("POSTGRES_PASSWORD", "password")
		PostgresDB := getEnv("POSTGRES_DB", "postgresdb")

		databaseURL = fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s",
			PostgresUser, PostgresPassword, PostgresHost, PostgresPort, PostgresDB,
		)
	}
	return databaseURL
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	if valueStr != "" {
		log.Printf("warning: invalid integer value %q for env %s, using default %d", valueStr, key, defaultValue)
	}
	return defaultValue
}
