package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Payment  PaymentConfig
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
	Port            int
	DB              int
}

type PaymentConfig struct {
	AlipayAppID    string
	AlipayPrivateKey string
	AlipayPublicKey string
	WechatAppID    string
	WechatMchID    string
	WechatAPIKey   string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{Port: getEnv("SERVER_PORT", "8083"), Mode: getEnv("GIN_MODE", "debug")},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"), Port: getEnv("DB_PORT", "3306"),
			User: getEnv("DB_USER", "root"), Password: getEnv("DB_PASSWORD", ""),
			DBName: getEnv("DB_NAME", "ecommerce_payment"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"), Port: 3,
			Password: getEnv("REDIS_PASSWORD", ""), DB: 3,
		},
		Payment: PaymentConfig{
			AlipayAppID:     getEnv("ALIPAY_APP_ID", ""),
			AlipayPrivateKey: getEnv("ALIPAY_PRIVATE_KEY", ""),
			AlipayPublicKey: getEnv("ALIPAY_PUBLIC_KEY", ""),
			WechatAppID:     getEnv("WECHAT_APP_ID", ""),
			WechatMchID:     getEnv("WECHAT_MCH_ID", ""),
			WechatAPIKey:    getEnv("WECHAT_API_KEY", ""),
		},
	}
}

func getEnv(k, v string) string {
	if val := os.Getenv(k); val != "" {
		return val
	}
	return v
}
