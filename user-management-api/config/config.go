package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI  string
	MongoDB   string
	JWTSecret string
	Port      string
}

// .env ไม่ทับ env var จริง (เช่นที่ Docker/CI inject มา) - env จริงชนะเสมอ
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	cfg := &Config{
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:   getEnv("MONGO_DB", "7solutions"),
		JWTSecret: getEnv("JWT_SECRET", "supersecret"),
		Port:      getEnv("PORT", "8080"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
