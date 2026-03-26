package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns the database connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	Provider string      `mapstructure:"provider"`
	Valkey   ValkeyConfig `mapstructure:"valkey"`
}

// ValkeyConfig holds Valkey connection configuration.
type ValkeyConfig struct {
	Addr     string `mapstructure:"addr"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
	Password string `mapstructure:"password"`
}

// StorageConfig holds storage backend configuration.
type StorageConfig struct {
	Provider  string `mapstructure:"provider"`
	Region    string `mapstructure:"region"`
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	GatewaySecret string         `mapstructure:"gateway_secret"`
	Keycloak      KeycloakConfig `mapstructure:"keycloak"`
	Session       SessionConfig  `mapstructure:"session"`
}

// KeycloakConfig holds Keycloak configuration.
type KeycloakConfig struct {
	BaseURL      string `mapstructure:"base_url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// SessionConfig holds session configuration.
type SessionConfig struct {
	TTL        int  `mapstructure:"ttl"`         // Session TTL in seconds
	Secure     bool `mapstructure:"secure"`      // Set Secure flag on cookie
	StateTTL   int  `mapstructure:"state_ttl"`   // OAuth state TTL in seconds
	RedisKey   string `mapstructure:"redis_key"` // Redis key prefix
}

// OIDCRedirectURI returns the OIDC redirect URI.
func (c *KeycloakConfig) OIDCRedirectURI(baseURL string) string {
	return baseURL + "/auth/callback"
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Set defaults
	v.SetDefault("app.name", "retrowin")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.env", "development")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("cache.provider", "valkey")
	v.SetDefault("cache.valkey.addr", "localhost:6379")
	v.SetDefault("cache.valkey.db", 0)
	v.SetDefault("cache.valkey.pool_size", 10)
	v.SetDefault("cache.valkey.password", "")
	v.SetDefault("storage.provider", "s3")
	v.SetDefault("storage.region", "us-east-1")
	v.SetDefault("storage.use_ssl", false)
	v.SetDefault("auth.session.ttl", 86400)         // 24 hours
	v.SetDefault("auth.session.secure", false)
	v.SetDefault("auth.session.state_ttl", 300)     // 5 minutes
	v.SetDefault("auth.session.redis_key", "retrowin")

	// Read environment variables
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
