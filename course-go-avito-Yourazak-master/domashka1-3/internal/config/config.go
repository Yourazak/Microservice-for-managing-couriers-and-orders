package config

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string            `json:"port"`
	DB              DBSettings        `json:"db"`
	ServiceOrderURL string            `json:"service_order_url"`
	Kafka           KafkaSettings     `json:"kafka"`
	Metrics         MetricsSettings   `json:"metrics"`
	RateLimit       RateLimitSettings `json:"rate_limit"`
	Pprof           PprofSettings     `json:"pprof"`
}

type DBSettings struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type KafkaSettings struct {
	Brokers       []string `json:"brokers"`
	OrderTopic    string   `json:"order_topic"`
	ConsumerGroup string   `json:"consumer_group"`
}

type MetricsSettings struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

type RateLimitSettings struct {
	Enabled           bool    `json:"enabled"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	Burst             int     `json:"burst"`
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	defaultPort := getEnv("PORT", "8080")
	flagPort := flag.String("port", defaultPort, "HTTP server port")
	flag.Parse()

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	metricsEnabled := getEnv("METRICS_ENABLED", "true") == "true"
	metricsPath := getEnv("METRICS_PATH", "/metrics")

	rateLimitEnabled := getEnv("RATE_LIMIT_ENABLED", "true") == "true"
	rateLimitRPS := parseFloat(getEnv("RATE_LIMIT_RPS", "5"))
	rateLimitBurst := parseInt(getEnv("RATE_LIMIT_BURST", "5"))

	pprofEnabled := getEnv("PPROF_ENABLED", "true") == "true"
	pprofPort := getEnv("PPROF_PORT", "6060")
	pprofEndpoint := getEnv("PPROF_ENDPOINT", "/debug/pprof")

	cfg := &Config{
		Port:            *flagPort,
		ServiceOrderURL: getEnv("SERVICE_ORDER_URL", "http://localhost:8080"),
		DB: DBSettings{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "myuser"),
			Password: getEnv("POSTGRES_PASSWORD", "mypassword"),
			Name:     getEnv("POSTGRES_DB", "test_db"),
		},
		Kafka: KafkaSettings{
			Brokers:       strings.Split(kafkaBrokers, ","),
			OrderTopic:    getEnv("KAFKA_ORDER_TOPIC", "order-events"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "courier-service"),
		},
		Metrics: MetricsSettings{
			Enabled: metricsEnabled,
			Path:    metricsPath,
		},
		RateLimit: RateLimitSettings{
			Enabled:           rateLimitEnabled,
			RequestsPerSecond: rateLimitRPS,
			Burst:             rateLimitBurst,
		},
		Pprof: PprofSettings{
			Enabled:  pprofEnabled,
			Port:     pprofPort,
			Endpoint: pprofEndpoint,
		},
	}

	validateConfig(cfg)
	return cfg
}

type PprofSettings struct {
	Enabled  bool   `json:"enabled"`
	Port     string `json:"port"`
	Endpoint string `json:"endpoint"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 5.0
	}
	return val
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 5
	}
	return val
}

func validateConfig(cfg *Config) {
	if cfg.Port == "" {
		panic("PORT is required")
	}
	if cfg.DB.Host == "" {
		panic("POSTGRES_HOST is required")
	}
	if cfg.DB.User == "" {
		panic("POSTGRES_USER is required")
	}
	if cfg.DB.Password == "" {
		panic("POSTGRES_PASSWORD is required")
	}
	if cfg.DB.Name == "" {
		panic("POSTGRES_DB is required")
	}
}
