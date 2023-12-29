package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/m-mizutani/zlog"
	"github.com/m-mizutani/zlog/filter"
)

type Config struct {
	Kafka          KafkaConfig          `mapstructure:"kafka" `
	SchemaRegistry SchemaRegistryConfig `mapstructure:"schemaRegistry"`
	Datagen        Datagen              `mapstructure:"datagen"`
}

type Datagen struct {
	Payments     int            `mapstructure:"payments"`
	Workers      int            `mapstructure:"workers"`
	Sources      int            `mapstructure:"sources"`
	Destinations int            `mapstructure:"destinations"`
	Workflows    map[string]int `mapstructure:"workflows"`
	Delays       map[string]int `mapstructure:"delays"`
}

type SchemaRegistryConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	ApiKey    string `mapstructure:"key"`
	ApiSecret string `mapstructure:"secret" zlog:"secret"`
}

type KafkaConfig struct {
	BootstrapServers string            `mapstructure:"bootstrapServers"`
	ClientId         string            `mapstructure:"clientId"`
	SaslMechanisms   string            `mapstructure:"saslMechanisms"`
	SecurityProtocol string            `mapstructure:"saslProtocol"`
	SaslUsername     string            `mapstructure:"saslUsername"`
	SaslPassword     string            `mapstructure:"saslPassword" zlog:"secret"`
	ConfigMap        map[string]string `mapstructure:"configMap"`
	Topics           map[string]int    `mapstructure:"topics"`
}

func Build() Config {

	logger := zlog.New(zlog.WithFilters(filter.Tag()))

	err := godotenv.Load() // 👈 load .env file
	if err != nil {
		log.Println(err)
	}

	config := Config{} // Init config
	config.Kafka.BootstrapServers = getenv("KAFKA_BOOTSTRAP_SERVER", "localhost:9092")
	config.Kafka.SaslUsername = getenv("KAFKA_SASL_USERNAME", "admin")
	config.Kafka.SaslPassword = getenv("KAFKA_SASL_PASSWORD", "admin-secret")
	config.Kafka.ConfigMap = map[string]string{
		"statistics.interval.ms":     "3000",
		"compression.codec":          "lz4",
		"batch.size":                 "648576",
		"linger.ms":                  "1000",
		"message.max.bytes":          "2000000",
		"queue.buffering.max.ms":     "1000",
		"queue.buffering.max.kbytes": "2048576",
		"batch.num.messages":         "10000",
	}
	config.Kafka.ClientId = getenv("KAFKA_CLIENT_ID", "synthethic-payment-generator")
	config.Kafka.SaslMechanisms = getenv("KAFKA_SASL_MECHANISMS", "PLAIN")
	config.Kafka.SecurityProtocol = getenv("KAFKA_SECURITY_PROTOCOL", "SASL_SSL")

	config.Kafka.Topics = map[string]int{
		"payment-initiated": 12,
		"payment-completed": 12,
		"payment-failed":    4,
		"payment-canceled":  4,
		"payment-validated": 12,
		"payment-accounted": 12,
		"payment-rejected":  4,
	}

	config.SchemaRegistry.Endpoint = getenv("SCHEMA_REGISTRY_ENDPOINT", "http://localhost:8081")
	config.SchemaRegistry.ApiKey = getenv("SCHEMA_REGISTRY_API_KEY", "")
	config.SchemaRegistry.ApiSecret = getenv("SCHEMA_REGISTRY_API_SECRET", "")

	config.Datagen.Payments = getenvInt("NUM_PAYMENTS", 100000)
	config.Datagen.Workers = getenvInt("NUM_WORKERS", 100)
	config.Datagen.Sources = getenvInt("NUM_SOURCES", 10)
	config.Datagen.Destinations = getenvInt("NUM_DESTINATIONS", 10)

	config.Datagen.Workflows = map[string]int{
		"Initiated, Failed":                          1,
		"Initiated, Rejected":                        2,
		"Initiated, Validated, Failed":               1,
		"Initiated, Validated, Rejected":             1,
		"Initiated, Validated, Accounted, Failed":    1,
		"Initiated, Validated, Accounted, Completed": 9,
		"Initiated, Validated, Accounted, Canceled":  2,
		"Initiated, Validated, Accounted, Rejected":  1,
	}

	config.Datagen.Delays = map[string]int{}

	config.Datagen.Delays["canceled"] = getenvInt("DELAY_CANCELED", 2000)
	config.Datagen.Delays["completed"] = getenvInt("DELAY_COMPLETED", 3000)
	config.Datagen.Delays["failed"] = getenvInt("DELAY_FAILED", 1000)
	config.Datagen.Delays["rejected"] = getenvInt("DELAY_REJECTED", 2000)
	config.Datagen.Delays["validated"] = getenvInt("DELAY_VALIDATED", 1000)
	config.Datagen.Delays["accounted"] = getenvInt("DELAY_ACCOUNTED", 1000)
	config.Datagen.Delays["initiated"] = getenvInt("DELAY_INITIATED", 100)

	logger.With("Config", config).Info("Configuration:")
	return config
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {
		return fallback
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	return value
}
