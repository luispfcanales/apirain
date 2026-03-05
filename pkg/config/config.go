package config

import (
	"os"
)

type Config struct {
	OdooURL         string
	OdooURLDev      string
	OdooDB          string
	OdooDBDev       string
	OdooUsername    string
	OdooPassword    string
	OdooPasswordDev string
	Port            string
}

func LoadConfig() (*Config, error) {
	return &Config{
		OdooUsername:    getEnv("ODOO_USER", "analista.ti@rainforestexpeditions.com"),
		OdooURL:         getEnv("ODOO_URL", "https://rainforest.odoo.com"),
		OdooDB:          getEnv("ODOO_DB", "alwaperugold-rainforest-main-22195173"),
		OdooPassword:    getEnv("ODOO_PWD", "1df02ad06707dae1ae1a1c655f7a08fbd70e3067"),
		OdooURLDev:      "https://rainforest-uat-ra-020326-29218273.dev.odoo.com",
		OdooDBDev:       "rainforest-uat-ra-020326-29218273",
		OdooPasswordDev: "6ee6c9c4c5daa71b77b3429c43a1814a8b5b4c23",
		Port:            getEnv("PORT", "8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
