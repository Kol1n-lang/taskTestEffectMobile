package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Configs struct {
	DB    DatabaseConfig
	Redis RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     string
	Database string
}

func Init() *Configs {
	if err := godotenv.Load(); err != nil {
		log.Println("Using static env variables")
	}
	config := &Configs{}

	config.DB = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		Username: getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "your_password"),
		Database: getEnv("DB_NAME", "your_database"),
	}

	config.Redis = RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Database: getEnv("REDIS_DB", "0"),
	}

	log.Printf("Config initialized")

	return config
}

func (dbSettings DatabaseConfig) DBUrl() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbSettings.Username,
		dbSettings.Password,
		dbSettings.Host,
		dbSettings.Port,
		dbSettings.Database,
	)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
