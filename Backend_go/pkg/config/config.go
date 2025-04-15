package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
}

type ServerConfig struct {
	Port    int           `mapstructure:"port" default:"8000"`
	Mode    string        `mapstructure:"mode"`
	Timeout time.Duration `mapstructure:"timeout"`
	UseHTTPS bool          `mapstructure:"use_https"`
	HTTPSCertFile string    `mapstructure:"https_cert_file"`
	HTTPSKeyFile  string    `mapstructure:"https_key_file"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslmode"`
	Timezone        string        `mapstructure:"timezone"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	JWTSecret       string                    `mapstructure:"jwt_secret"`
	JWTExpiryHours  int                       `mapstructure:"jwt_expiry_hours"`
	JWTIssuer       string                    `mapstructure:"jwt_issuer"`
	OAuth2          OAuth2Config              `mapstructure:"oauth2"`
	OAuth2Providers map[string]ProviderConfig `mapstructure:"oauth2_providers"`
}

type OAuth2Config struct {
	Enabled      bool   `mapstructure:"enabled"`
	CallbackURL  string `mapstructure:"callback_url"`
	StateTimeout int    `mapstructure:"state_timeout"` // in minutes
}

type ProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	UserInfoURL  string   `mapstructure:"userinfo_url"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
	Host        string `mapstructure:"host"`
	BasePath    string `mapstructure:"base_path"`
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// If CONFIG_FILE environment variable is set, use it
	if envConfigFile := os.Getenv("CONFIG_FILE"); envConfigFile != "" {
		configPath = envConfigFile
	}

	// Initialize viper
	v := viper.New()
	v.SetConfigType("yaml")

	// If configPath is provided, use it directly
	if configPath != "" {
		// Get the directory and filename
		dir := filepath.Dir(configPath)
		file := filepath.Base(configPath)
		ext := filepath.Ext(file)
		name := strings.TrimSuffix(file, ext)

		v.AddConfigPath(dir)
		v.SetConfigName(name)
	} else {
		// Fallback to default locations
		_, filename, _, _ := runtime.Caller(0)
		pkgConfigDir := filepath.Dir(filename)
		projectRoot := filepath.Join(pkgConfigDir, "..", "..")

		v.AddConfigPath(pkgConfigDir)
		v.AddConfigPath(projectRoot)
		v.AddConfigPath(filepath.Join(projectRoot, "pkg", "config"))
		v.SetConfigName("config")
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}

	// Enable environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Override with environment variables if they exist
	envVars := map[string]string{
		"database.host":                          "DB_HOST",
		"database.port":                          "DB_PORT",
		"database.user":                          "DB_USER",
		"database.password":                      "DB_PASSWORD",
		"database.name":                          "DB_NAME",
		"database.sslmode":                       "DB_SSLMODE",
		"server.mode":                            "SERVER_MODE",
		"server.timeout":                         "SERVER_TIMEOUT",
		"redis.host":                             "REDIS_HOST",
		"redis.port":                             "REDIS_PORT",
		"redis.password":                         "REDIS_PASSWORD",
		"redis.db":                               "REDIS_DB",
		"auth.jwt_secret":                        "JWT_SECRET",
		"auth.jwt_issuer":                        "JWT_ISSUER",
		"auth.jwt_expiry_hours":                  "JWT_EXPIRY_HOURS",
		"auth.oauth2.enabled":                    "OAUTH2_ENABLED",
		"auth.oauth2.callback_url":               "OAUTH2_CALLBACK_URL",
		"auth.oauth2.state_timeout":              "OAUTH2_STATE_TIMEOUT",
		"auth.oauth2_providers.google.client_id": "OAUTH2_GOOGLE_CLIENT_ID",
		"auth.oauth2_providers.google.client_secret": "OAUTH2_GOOGLE_CLIENT_SECRET",
		"auth.oauth2_providers.google.redirect_url":  "OAUTH2_GOOGLE_REDIRECT_URL",
		"auth.oauth2_providers.github.client_id":     "OAUTH2_GITHUB_CLIENT_ID",
		"auth.oauth2_providers.github.client_secret": "OAUTH2_GITHUB_CLIENT_SECRET",
		"auth.oauth2_providers.github.redirect_url":  "OAUTH2_GITHUB_REDIRECT_URL",
		"logging.level":  "LOG_LEVEL",
		"logging.format": "LOG_FORMAT",
	}

	for configKey, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			// Handle special cases for type conversion
			switch envVar {
			case "DB_PORT", "REDIS_PORT", "JWT_EXPIRY_HOURS", "OAUTH2_STATE_TIMEOUT":
				if intVal, err := strconv.Atoi(value); err == nil {
					v.Set(configKey, intVal)
				}
			case "SERVER_TIMEOUT":
				if d, err := time.ParseDuration(value); err == nil {
					v.Set(configKey, d)
				}
			case "OAUTH2_ENABLED":
				if value == "true" || value == "1" {
					v.Set(configKey, true)
				} else if value == "false" || value == "0" {
					v.Set(configKey, false)
				}
			default:
				v.Set(configKey, value)
			}
		}
	}

	// Unmarshal config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}
