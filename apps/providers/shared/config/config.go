package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Load loads local configuration and checks required service credentials.
func Load() error {
	path := Get("ENV_FILE", ".env")
	if err := godotenv.Load(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read environment file")
	}
	if len(os.Getenv("INTERNAL_API_KEY")) < 32 {
		return fmt.Errorf("INTERNAL_API_KEY must contain at least 32 characters")
	}
	return nil
}

// Get returns a configuration value or its fallback.
func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ServiceToken derives a caller-specific credential for internal HTTP routes.
func ServiceToken(service string) string {
	mac := hmac.New(sha256.New, []byte(os.Getenv("INTERNAL_API_KEY")))
	_, _ = mac.Write([]byte(service))
	return hex.EncodeToString(mac.Sum(nil))
}
