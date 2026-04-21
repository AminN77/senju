// Package config loads process configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime settings for the API process and readiness wiring.
type Config struct {
	APIPort int
	// LogLevel is a zerolog level name (debug, info, warn, error). Empty defaults to info.
	LogLevel string

	PostgresDSN    string
	ClickHousePing string
	ClickHouseDSN  string
	MinIOHealthURL string
	NATSAddr       string

	// ObjectStore configures S3-compatible multipart uploads (e.g. MinIO). See ObjectStoreConfig.Enabled.
	ObjectStore ObjectStoreConfig
	Queue       QueueConfig
}

// ObjectStoreConfig holds S3-compatible API settings for presigned multipart uploads.
// Enabled when Endpoint, Bucket, AccessKey, and SecretKey are all non-empty.
type ObjectStoreConfig struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

// QueueConfig defines NATS/JetStream queue behavior for worker retries.
type QueueConfig struct {
	StreamName   string
	Subject      string
	DeadLetter   string
	ConsumerName string
	MaxRetries   int
	BackoffBase  time.Duration
}

// Enabled reports whether object multipart routes should call the object store (vs 503).
func (o ObjectStoreConfig) Enabled() bool {
	return o.Endpoint != "" && o.Bucket != "" && o.AccessKey != "" && o.SecretKey != ""
}

// Load reads configuration from the environment.
func Load() (Config, error) {
	portStr := os.Getenv("API_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil {
		return Config{}, fmt.Errorf("API_PORT: parse: %w", err)
	}
	if port < 1 || port > 65535 {
		return Config{}, errors.New("API_PORT: must be a decimal integer between 1 and 65535")
	}

	chBase := strings.TrimRight(os.Getenv("CLICKHOUSE_HTTP_URL"), "/")
	var chPing string
	if chBase != "" {
		chPing = chBase + "/ping"
	}
	chDSN, err := clickHouseDSN(chBase)
	if err != nil {
		return Config{}, err
	}

	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}

	usePathStyle := true
	if v := strings.TrimSpace(os.Getenv("S3_USE_PATH_STYLE")); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			usePathStyle = b
		}
	}

	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY"))
	if accessKey == "" {
		accessKey = strings.TrimSpace(os.Getenv("MINIO_ROOT_USER"))
	}
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_KEY"))
	if secretKey == "" {
		secretKey = strings.TrimSpace(os.Getenv("MINIO_ROOT_PASSWORD"))
	}

	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("MINIO_S3_ENDPOINT"))
	}

	queueMaxRetries, err := parseIntEnv("QUEUE_MAX_RETRIES", 3)
	if err != nil {
		return Config{}, err
	}
	queueBackoffBase, err := parseDurationEnv("QUEUE_BACKOFF_BASE", 1*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		APIPort:        port,
		LogLevel:       os.Getenv("LOG_LEVEL"),
		PostgresDSN:    postgresDSN(),
		ClickHousePing: chPing,
		ClickHouseDSN:  chDSN,
		MinIOHealthURL: strings.TrimSpace(os.Getenv("MINIO_HEALTH_URL")),
		NATSAddr:       strings.TrimSpace(os.Getenv("NATS_ADDR")),
		ObjectStore: ObjectStoreConfig{
			Endpoint:     endpoint,
			Region:       region,
			Bucket:       strings.TrimSpace(os.Getenv("S3_BUCKET")),
			AccessKey:    accessKey,
			SecretKey:    secretKey,
			UsePathStyle: usePathStyle,
		},
		Queue: QueueConfig{
			StreamName:   getenvDefaultTrim("QUEUE_STREAM_NAME", "jobs_stream"),
			Subject:      getenvDefaultTrim("QUEUE_SUBJECT", "jobs.execute"),
			DeadLetter:   getenvDefaultTrim("QUEUE_DEAD_LETTER_SUBJECT", "jobs.dead_letter"),
			ConsumerName: getenvDefaultTrim("QUEUE_CONSUMER_NAME", "jobs_worker"),
			MaxRetries:   queueMaxRetries,
			BackoffBase:  queueBackoffBase,
		},
	}, nil
}

func clickHouseDSN(chBase string) (string, error) {
	if dsn := strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN")); dsn != "" {
		return dsn, nil
	}
	if chBase == "" {
		return "", nil
	}
	u, err := url.Parse(chBase)
	if err != nil {
		return "", fmt.Errorf("CLICKHOUSE_HTTP_URL: parse: %w", err)
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return "", errors.New("CLICKHOUSE_HTTP_URL: hostname is required")
	}
	port := getenvDefaultTrim("CLICKHOUSE_NATIVE_PORT", "9000")
	user := getenvDefaultTrim("CLICKHOUSE_USER", "default")
	pass := strings.TrimSpace(os.Getenv("CLICKHOUSE_PASSWORD"))
	db := getenvDefaultTrim("CLICKHOUSE_DB", "default")

	out := &url.URL{
		Scheme: "clickhouse",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + db,
	}
	return out.String(), nil
}

func postgresDSN() string {
	if d := strings.TrimSpace(os.Getenv("POSTGRES_DSN")); d != "" {
		return d
	}
	host := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
	if host == "" {
		return ""
	}
	port := getenvDefault("POSTGRES_PORT", "5432")
	user := getenvDefault("POSTGRES_USER", "senju")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := getenvDefault("POSTGRES_DB", "senju")

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbname,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvDefaultTrim(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v != "" {
		return v
	}
	return def
}

func parseIntEnv(key string, def int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def, nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse int: %w", key, err)
	}
	return i, nil
}

func parseDurationEnv(key string, def time.Duration) (time.Duration, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse duration: %w", key, err)
	}
	return d, nil
}
