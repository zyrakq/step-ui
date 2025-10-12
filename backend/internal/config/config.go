package config

import (
	"os"
	"strconv"
)

type Config struct {
	CAURL                   string
	CARoot                  string
	ProvisionerName         string
	ProvisionerPasswordFile string
	DBPath                  string
	Port                    int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	
	return &Config{
		CAURL:                   getEnv("CA_URL", ""),
		CARoot:                  getEnv("CA_ROOT", ""),
		ProvisionerName:         getEnv("PROVISIONER_NAME", "ui-admin"),
		ProvisionerPasswordFile: getEnv("PROVISIONER_PASSWORD_FILE", ""),
		DBPath:                  getEnv("DB_PATH", "./data/certs.db"),
		Port:                    port,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
