package config

import (
	"os"
	"strconv"
)

type Config struct {
	CAURL             string
	ProvisionerName   string
	ProvisionerPassword string
	DBPath            string
	Port              int
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	
	return &Config{
		CAURL:             getEnv("CA_URL", ""),
		ProvisionerName:   getEnv("PROVISIONER_NAME", "ui-admin"),
		ProvisionerPassword: getEnv("PROVISIONER_PASSWORD", ""),
		DBPath:            getEnv("DB_PATH", "./data/certs.db"),
		Port:              port,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
