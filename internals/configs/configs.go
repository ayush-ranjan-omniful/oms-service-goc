package configs

import (
	"os"

	"github.com/omniful/go_commons/db/nosql/mongodm"
	"github.com/omniful/go_commons/sqs"
)

type Config struct {
	Server  ServerConfig   `json:"server"`
	MongoDB mongodm.Config `json:"mongodb"`
	SQS     *sqs.Config    `json:"sqs"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

func LoadConfig() *Config {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	if env == "local" {
		return &Config{
			Server: ServerConfig{
				Port: ":8080",
			},
			MongoDB: mongodm.Config{
				Database: "oms_db",
				URI:      "mongodb://localhost:27017",
			},
			SQS: &sqs.Config{
				Account:  "000000000000",
				Region:   "us-east-1",
				Endpoint: "http://localhost:4566", // LocalStack endpoint
			},
		}
	}

	// Production: use environment variables
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", ":8080"),
		},
		MongoDB: mongodm.Config{
			Database: getEnv("MONGODB_DATABASE", "oms_db"),
			URI:      os.Getenv("MONGODB_URI"),
		},
		SQS: &sqs.Config{
			Account:  os.Getenv("SQS_ACCOUNT"),
			Region:   os.Getenv("SQS_REGION"),
			Endpoint: os.Getenv("SQS_ENDPOINT"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
