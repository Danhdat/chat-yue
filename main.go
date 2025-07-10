package main

import (
	"chatbtc/config"
	"chatbtc/models"
	"chatbtc/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	http.HandleFunc("/health", healthCheck)
	go func() {
		log.Println("Healthcheck server running at :8080/health")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	// Load cấu hình
	config.LoadConfig()

	// Kiểm tra các biến môi trường bắt buộc
	if config.AppConfig.TelegramBotToken == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN không được cấu hình")
	}

	log.Println("🚀 Khởi động Crypto Analysis Bot...")

	// Khởi tạo database
	if err := models.InitDatabase(); err != nil {
		log.Fatalf("❌ Lỗi kết nối database: %v", err)
	}
	defer models.CloseDatabase()

	// Auto migrate database
	if err := models.AutoMigrate(); err != nil {
		log.Fatalf("❌ Lỗi migrate database: %v", err)
	}

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

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
