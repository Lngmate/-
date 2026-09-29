package config

import (
	"os"
)

type Config struct {
	Port        string
	DBType      string // "sqlite" or "mysql"
	DSN         string
	JWTSecret   string
	RedisAddr   string
}

var AppConfig *Config

func InitConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite" // 默认使用纯Go无依赖SQLite，即开即用
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		if dbType == "mysql" {
			dsn = "root:root@tcp(127.0.0.1:3306)/lab_system?charset=utf8mb4&parseTime=True&loc=Local"
		} else {
			dsn = "lab_system.db"
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "lab_reservation_secret_key_2026"
	}

	AppConfig = &Config{
		Port:      port,
		DBType:    dbType,
		DSN:       dsn,
		JWTSecret: jwtSecret,
		RedisAddr: os.Getenv("REDIS_ADDR"),
	}

	return AppConfig
}
