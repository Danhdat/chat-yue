package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	ProxyEnabled     bool
	ProxyType        string
	ProxyURL         string
	ProxyUsername    string
	ProxyPassword    string
	CryptoAPIKey     string
	CryptoAPIURL     string
	ServerPort       string
	LogLevel         string
	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPassword       string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	proxyURL := getEnv("PROXY_URL", "")
	// Tự động thêm protocol nếu chưa có (chỉ cho http)
	proxyType := strings.ToLower(getEnv("PROXY_TYPE", "http"))
	if proxyType != "http" && proxyType != "https" && proxyType != "socks5" {
		proxyType = "http" // Mặc định về HTTP nếu type không hợp lệ
	}

	if proxyType == "http" && proxyURL != "" && !strings.HasPrefix(proxyURL, "http://") {
		proxyURL = "http://" + proxyURL
	} else if proxyType == "https" && proxyURL != "" && !strings.HasPrefix(proxyURL, "https://") {
		proxyURL = "https://" + proxyURL
	}

	AppConfig = &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		ProxyEnabled:     getEnvAsBool("PROXY_ENABLED", false),
		ProxyType:        proxyType,
		ProxyURL:         proxyURL,
		ProxyUsername:    getEnv("PROXY_USERNAME", ""),
		ProxyPassword:    getEnv("PROXY_PASSWORD", ""),
		CryptoAPIKey:     getEnv("CRYPTO_API_KEY", ""),
		CryptoAPIURL:     getEnv("CRYPTO_API_URL", "https://api.coingecko.com/api/v3"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		DBHost:           getEnv("DB_HOST", ""),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBName:           getEnv("DB_NAME", "cryptobot"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
