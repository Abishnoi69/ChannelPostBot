package config

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Token       = getEnv("TOKEN", "")
	OwnerId     = getEnvInt64("OWNER_ID", 5938660179)
	LoggerId    = getEnvInt64("LOGGER_ID", 5938660179)
	DatabaseURI = getEnv("DB_URI", "")
	DbName      = getEnv("DB_NAME", "ChannelManager")
	RedisURI    = getEnv("REDIS_URI", "")
	SupportChat = getEnv("SUPPORT_CHAT", "")

	WebhookUrl = getEnv("WEBHOOK_URL", "")
	Port       = getEnv("PORT", "9099")
)

// getEnv returns the value of an environment variable or a default value if it is not set
func getEnv(key, value string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return value
}

// getEnvInt64 returns the value of an environment variable as an int64 or a default value if it is not set
func getEnvInt64(key string, value int64) int64 {
	if value, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil {
		return value
	}
	return value
}

var (
	FakeDevs = []int64{5938660179, 241413457, 885488992, 287831937, 6366452257}
)
