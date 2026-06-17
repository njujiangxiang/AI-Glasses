// Package config 集中管理后端运行配置。配置来源依次为代码默认值、config.yaml 等 YAML
// 配置文件和环境变量覆盖，从而让本地开发与部署环境共用一套可预期的加载流程。
package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr              string        `yaml:"http_addr"`
	DatabaseDSN           string        `yaml:"database_dsn"`
	JWTSecret             string        `yaml:"jwt_secret"`
	AccessTokenTTL        time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL       time.Duration `yaml:"refresh_token_ttl"`
	RedisAddr             string        `yaml:"redis_addr"`
	RedisPassword         string        `yaml:"redis_password"`
	RabbitMQURL           string        `yaml:"rabbitmq_url"`
	S3Endpoint            string        `yaml:"s3_endpoint"`
	S3AccessKey           string        `yaml:"s3_access_key"`
	S3SecretKey           string        `yaml:"s3_secret_key"`
	S3Bucket              string        `yaml:"s3_bucket"`
	S3UseSSL              bool          `yaml:"s3_use_ssl"`
	SchedulerLookback     time.Duration `yaml:"scheduler_lookback"`
	RequiredPhotoMaxBytes int64         `yaml:"required_photo_max_bytes"`
	AudioMaxBytes         int64         `yaml:"audio_max_bytes"`
}

type fileConfig struct {
	HTTPAddr              string `yaml:"http_addr"`
	DatabaseDSN           string `yaml:"database_dsn"`
	JWTSecret             string `yaml:"jwt_secret"`
	AccessTokenTTL        string `yaml:"access_token_ttl"`
	RefreshTokenTTL       string `yaml:"refresh_token_ttl"`
	RedisAddr             string `yaml:"redis_addr"`
	RedisPassword         string `yaml:"redis_password"`
	RabbitMQURL           string `yaml:"rabbitmq_url"`
	S3Endpoint            string `yaml:"s3_endpoint"`
	S3AccessKey           string `yaml:"s3_access_key"`
	S3SecretKey           string `yaml:"s3_secret_key"`
	S3Bucket              string `yaml:"s3_bucket"`
	S3UseSSL              *bool  `yaml:"s3_use_ssl"`
	SchedulerLookback     string `yaml:"scheduler_lookback"`
	RequiredPhotoMaxBytes *int64 `yaml:"required_photo_max_bytes"`
	AudioMaxBytes         *int64 `yaml:"audio_max_bytes"`
}

// Load 加载完整后端配置，加载顺序为默认值、配置文件、环境变量覆盖。
func Load() Config {
	cfg := Config{
		HTTPAddr:              ":8080",
		DatabaseDSN:           "aiglasses:aiglasses@tcp(127.0.0.1:3306)/aiglasses?charset=utf8mb4&parseTime=True&loc=UTC",
		JWTSecret:             "dev-only-change-me",
		AccessTokenTTL:        30 * time.Minute,
		RefreshTokenTTL:       30 * 24 * time.Hour,
		RedisAddr:             "",
		RedisPassword:         "",
		RabbitMQURL:           "",
		S3Endpoint:            "",
		S3AccessKey:           "",
		S3SecretKey:           "",
		S3Bucket:              "",
		S3UseSSL:              false,
		SchedulerLookback:     24 * time.Hour,
		RequiredPhotoMaxBytes: 10 * 1024 * 1024,
		AudioMaxBytes:         30 * 1024 * 1024,
	}
	applyFile(&cfg, env("CONFIG_FILE", "config.yaml"))
	applyEnv(&cfg)
	return cfg
}

// applyFile 从 YAML 配置文件读取配置，并把存在的字段覆盖到当前配置对象。
func applyFile(cfg *Config, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var file fileConfig
	if err := yaml.Unmarshal(data, &file); err != nil {
		return
	}
	if file.HTTPAddr != "" {
		cfg.HTTPAddr = file.HTTPAddr
	}
	if file.DatabaseDSN != "" {
		cfg.DatabaseDSN = file.DatabaseDSN
	}
	if file.JWTSecret != "" {
		cfg.JWTSecret = file.JWTSecret
	}
	if parsed, ok := parseDuration(file.AccessTokenTTL); ok {
		cfg.AccessTokenTTL = parsed
	}
	if parsed, ok := parseDuration(file.RefreshTokenTTL); ok {
		cfg.RefreshTokenTTL = parsed
	}
	if file.RedisAddr != "" {
		cfg.RedisAddr = file.RedisAddr
	}
	if file.RedisPassword != "" {
		cfg.RedisPassword = file.RedisPassword
	}
	if file.RabbitMQURL != "" {
		cfg.RabbitMQURL = file.RabbitMQURL
	}
	if file.S3Endpoint != "" {
		cfg.S3Endpoint = file.S3Endpoint
	}
	if file.S3AccessKey != "" {
		cfg.S3AccessKey = file.S3AccessKey
	}
	if file.S3SecretKey != "" {
		cfg.S3SecretKey = file.S3SecretKey
	}
	if file.S3Bucket != "" {
		cfg.S3Bucket = file.S3Bucket
	}
	if file.S3UseSSL != nil {
		cfg.S3UseSSL = *file.S3UseSSL
	}
	if parsed, ok := parseDuration(file.SchedulerLookback); ok {
		cfg.SchedulerLookback = parsed
	}
	if file.RequiredPhotoMaxBytes != nil {
		cfg.RequiredPhotoMaxBytes = *file.RequiredPhotoMaxBytes
	}
	if file.AudioMaxBytes != nil {
		cfg.AudioMaxBytes = *file.AudioMaxBytes
	}
}

// applyEnv 使用环境变量覆盖配置文件和默认值，便于部署或临时调试。
func applyEnv(cfg *Config) {
	cfg.HTTPAddr = env("HTTP_ADDR", cfg.HTTPAddr)
	cfg.DatabaseDSN = env("DATABASE_DSN", cfg.DatabaseDSN)
	cfg.JWTSecret = env("JWT_SECRET", cfg.JWTSecret)
	cfg.AccessTokenTTL = durationEnv("ACCESS_TOKEN_TTL", cfg.AccessTokenTTL)
	cfg.RefreshTokenTTL = durationEnv("REFRESH_TOKEN_TTL", cfg.RefreshTokenTTL)
	cfg.RedisAddr = env("REDIS_ADDR", cfg.RedisAddr)
	cfg.RedisPassword = env("REDIS_PASSWORD", cfg.RedisPassword)
	cfg.RabbitMQURL = env("RABBITMQ_URL", cfg.RabbitMQURL)
	cfg.S3Endpoint = env("S3_ENDPOINT", cfg.S3Endpoint)
	cfg.S3AccessKey = env("S3_ACCESS_KEY", cfg.S3AccessKey)
	cfg.S3SecretKey = env("S3_SECRET_KEY", cfg.S3SecretKey)
	cfg.S3Bucket = env("S3_BUCKET", cfg.S3Bucket)
	cfg.S3UseSSL = boolEnv("S3_USE_SSL", cfg.S3UseSSL)
	cfg.SchedulerLookback = durationEnv("SCHEDULER_LOOKBACK", cfg.SchedulerLookback)
	cfg.RequiredPhotoMaxBytes = int64Env("REQUIRED_PHOTO_MAX_BYTES", cfg.RequiredPhotoMaxBytes)
	cfg.AudioMaxBytes = int64Env("AUDIO_MAX_BYTES", cfg.AudioMaxBytes)
}

// env 读取字符串环境变量，不存在时返回传入的默认值。
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// durationEnv 读取 duration 类型环境变量，解析失败时保留默认值。
func durationEnv(key string, fallback time.Duration) time.Duration {
	parsed, ok := parseDuration(os.Getenv(key))
	if !ok {
		return fallback
	}
	return parsed
}

// parseDuration 将字符串解析为 time.Duration，并用布尔值标识解析是否成功。
func parseDuration(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// boolEnv 读取布尔环境变量，解析失败时返回默认值。
func boolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// int64Env 读取 int64 环境变量，解析失败时返回默认值。
func int64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
