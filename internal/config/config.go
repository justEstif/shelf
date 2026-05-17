package config

import (
	"os"
	"strconv"
)

type Config struct {
	Password       string
	ViewerPassword string
	Port           string
	PagesDir       string
	DataDir        string
	BaseURL        string
	CSRFKey        string
	MaxFileSize    int64 // bytes, default 500 MB
	MaxVolumeSize  int64 // bytes, default 1 GB
}

func Load() *Config {
	return &Config{
		Password:       os.Getenv("SHELF_PASSWORD"),
		ViewerPassword: os.Getenv("SHELF_VIEWER_PASSWORD"),
		Port:           getEnv("SHELF_PORT", getEnv("PORT", "3000")),
		PagesDir:       getEnv("SHELF_PAGES_DIR", "./pages"),
		DataDir:        getEnv("SHELF_DATA_DIR", "./data"),
		BaseURL:        getEnv("SHELF_BASE_URL", "http://localhost:3000"),
		CSRFKey:        os.Getenv("CSRF_KEY"),
		MaxFileSize:    getEnvInt("SHELF_MAX_FILE_SIZE", 500),    // MB
		MaxVolumeSize:  getEnvInt("SHELF_MAX_VOLUME_SIZE", 1024), // MB
	}
}

const mb = 1024 * 1024

func (c *Config) MaxFileSizeBytes() int64    { return c.MaxFileSize * mb }
func (c *Config) MaxVolumeSizeBytes() int64  { return c.MaxVolumeSize * mb }

func (c *Config) IsProduction() bool {
	return c.BaseURL != "" && c.BaseURL != "http://localhost:3000"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
