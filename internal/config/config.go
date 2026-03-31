package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	App      AppConfig      `mapstructure:"app" yaml:"app"`
	HTTP     HTTPConfig     `mapstructure:"http" yaml:"http"`
	Database DatabaseConfig `mapstructure:"database" yaml:"database"`
	Cache    CacheConfig    `mapstructure:"cache" yaml:"cache"`
	Storage  StorageConfig  `mapstructure:"storage" yaml:"storage"`
	Auth     AuthConfig     `mapstructure:"auth" yaml:"auth"`
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name    string `mapstructure:"name" yaml:"name"`
	Version string `mapstructure:"version" yaml:"version"`
	Env     string `mapstructure:"env" yaml:"env"`
}

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Host        string `mapstructure:"host" yaml:"host"`
	Port        int    `mapstructure:"port" yaml:"port"`
	OpenAPIPath string `mapstructure:"openapi_path" yaml:"openapi_path"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver" yaml:"driver"`
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Name     string `mapstructure:"name" yaml:"name"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
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
	Provider string       `mapstructure:"provider" yaml:"provider"`
	Valkey   ValkeyConfig `mapstructure:"valkey" yaml:"valkey"`
}

// ValkeyConfig holds Valkey connection configuration.
type ValkeyConfig struct {
	Addr     string `mapstructure:"addr" yaml:"addr"`
	DB       int    `mapstructure:"db" yaml:"db"`
	PoolSize int    `mapstructure:"pool_size" yaml:"pool_size"`
	Password string `mapstructure:"password" yaml:"password"`
}

// StorageConfig holds storage backend configuration.
type StorageConfig struct {
	Provider  string `mapstructure:"provider" yaml:"provider"`
	Region    string `mapstructure:"region" yaml:"region"`
	Endpoint  string `mapstructure:"endpoint" yaml:"endpoint"`
	AccessKey string `mapstructure:"access_key" yaml:"access_key"`
	SecretKey string `mapstructure:"secret_key" yaml:"secret_key"`
	Bucket    string `mapstructure:"bucket" yaml:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Keycloak KeycloakConfig `mapstructure:"keycloak" yaml:"keycloak"`
	Session  SessionConfig  `mapstructure:"session" yaml:"session"`
}

// KeycloakConfig holds Keycloak configuration.
type KeycloakConfig struct {
	BaseURL      string `mapstructure:"base_url" yaml:"base_url"`
	Realm        string `mapstructure:"realm" yaml:"realm"`
	ClientID     string `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string `mapstructure:"client_secret" yaml:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri" yaml:"redirect_uri"`
}

// SessionConfig holds session configuration.
type SessionConfig struct {
	TTL      int    `mapstructure:"ttl" yaml:"ttl"`             // Session TTL in seconds
	Secure   bool   `mapstructure:"secure" yaml:"secure"`       // Set Secure flag on cookie
	StateTTL int    `mapstructure:"state_ttl" yaml:"state_ttl"` // OAuth state TTL in seconds
	RedisKey string `mapstructure:"redis_key" yaml:"redis_key"` // Redis key prefix
}

// OIDCRedirectURI returns the OIDC redirect URI.
func (c *KeycloakConfig) OIDCRedirectURI(baseURL string) string {
	return baseURL + "/auth/callback"
}

// DSN returns the database connection string.
func (c *Config) DSN() string {
	return c.Database.DSN()
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Env) == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Env) == "production"
}

// Load reads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)
	bindEnvVars(v)

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadFromPath reads configuration from a specific path.
func LoadFromPath(configPath string) (*Config, error) {
	return Load(configPath)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "retrowin")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.env", "development")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.openapi_path", "api/openapi.yaml")
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
	v.SetDefault("auth.session.ttl", 86400) // 24 hours
	v.SetDefault("auth.session.secure", false)
	v.SetDefault("auth.session.state_ttl", 300) // 5 minutes
	v.SetDefault("auth.session.redis_key", "retrowin")
	v.SetDefault("auth.keycloak.redirect_uri", "http://localhost:8080/auth/callback")
}

func bindEnvVars(v *viper.Viper) {
	// Support nested env vars like DATABASE_PASSWORD, CACHE_VALKEY_PASSWORD
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind common secrets
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD")
	_ = v.BindEnv("cache.valkey.password", "CACHE_VALKEY_PASSWORD")
	_ = v.BindEnv("storage.access_key", "STORAGE_ACCESS_KEY")
	_ = v.BindEnv("storage.secret_key", "STORAGE_SECRET_KEY")
	_ = v.BindEnv("auth.keycloak.client_secret", "AUTH_KEYCLOAK_CLIENT_SECRET")
}
