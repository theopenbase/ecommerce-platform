package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	SMS      SMSConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int64
	RefreshTokenTTL int64
}

type SMSConfig struct {
	AccessKey  string
	SecretKey  string
	SignName   string
	TemplateID string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "ecommerce_user"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-256bit-secret"),
			AccessTokenTTL:  900,   // 15 min
			RefreshTokenTTL: 604800, // 7 days
		},
		SMS: SMSConfig{
			AccessKey:  getEnv("SMS_ACCESS_KEY", ""),
			SecretKey:  getEnv("SMS_SECRET_KEY", ""),
			SignName:   getEnv("SMS_SIGN_NAME", "电商平台"),
			TemplateID: getEnv("SMS_TEMPLATE_ID", "SMS_123456789"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
