package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host, Port, User, Password, DBName string
}

type RedisConfig struct {
	Host, Password string
	Port           int
	DB             int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{Port: getEnv("SERVER_PORT", "8084"), Mode: getEnv("GIN_MODE", "debug")},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "3306"),
			User: getEnv("DB_USER", "root"), Password: getEnv("DB_PASSWORD", ""),
			DBName: getEnv("DB_NAME", "ecommerce_marketing"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"), Port: 4,
			Password: getEnv("REDIS_PASSWORD", ""), DB: 4,
		},
	}
}

func getEnv(k, v string) string {
	if val := os.Getenv(k); val != "" {
		return val
	}
	return v
}
