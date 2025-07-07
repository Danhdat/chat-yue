package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"chatbtc/config"
	"chatbtc/services"
)

func main() {
	// Load cấu hình
	config.LoadConfig()

	// Kiểm tra các biến môi trường bắt buộc
	if config.AppConfig.TelegramBotToken == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN không được cấu hình")
	}

	log.Println("🚀 Khởi động Crypto Analysis Bot...")

	// Khởi tạo Telegram bot service
	botService, err := services.NewTelegramBotService()
	if err != nil {
		log.Fatalf("❌ Lỗi khởi tạo bot: %v", err)
	}

	// Tạo channel để nhận tín hiệu dừng
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Khởi động bot trong goroutine
	go func() {
		log.Println("✅ Bot đã sẵn sàng nhận tin nhắn...")
		botService.StartBot()
	}()

	// Chờ tín hiệu dừng
	<-stopChan
	log.Println("🛑 Đang dừng bot...")
}
