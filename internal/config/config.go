package config

import (
	"os"
)

type Config struct {
	Password       string
	ViewerPassword string
	Port           string
	PagesDir       string
	DataDir        string
	BaseURL        string
	CSRFKey        string
}

func Load() *Config {
	return &Config{
		Password:       os.Getenv("SHELF_PASSWORD"),
		ViewerPassword: os.Getenv("SHELF_VIEWER_PASSWORD"),
		Port:           getEnv("SHELF_PORT", "3000"),
		PagesDir:       getEnv("SHELF_PAGES_DIR", "./pages"),
		DataDir:        getEnv("SHELF_DATA_DIR", "./data"),
		BaseURL:        getEnv("SHELF_BASE_URL", "http://localhost:3000"),
		CSRFKey:        os.Getenv("CSRF_KEY"),
	}
}

func (c *Config) IsProduction() bool {
	return c.BaseURL != "" && c.BaseURL != "http://localhost:3000"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
