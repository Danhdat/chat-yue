.PHONY: help build run test clean deps setup

# Default target
help:
	@echo "🚀 Crypto Analysis Telegram Bot - Makefile"
	@echo ""
	@echo "Các lệnh có sẵn:"
	@echo "  setup    - Cài đặt dependencies và cấu hình ban đầu"
	@echo "  build    - Build ứng dụng"
	@echo "  run      - Chạy ứng dụng"
	@echo "  test     - Chạy tests"
	@echo "  clean    - Xóa files build"
	@echo "  deps     - Cài đặt dependencies"

# Cài đặt ban đầu
setup: deps
	@echo "📝 Tạo file cấu hình..."
	@if [ ! -f .env ]; then \
		cp config.env.example .env; \
		echo "✅ Đã tạo file .env từ config.env.example"; \
		echo "⚠️  Vui lòng chỉnh sửa file .env với thông tin của bạn"; \
	else \
		echo "✅ File .env đã tồn tại"; \
	fi

# Cài đặt dependencies
deps:
	@echo "📦 Cài đặt dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies đã được cài đặt"

# Build ứng dụng
build:
	@echo "🔨 Building ứng dụng..."
	go build -o bin/cryptobot main.go
	@echo "✅ Build hoàn thành: bin/cryptobot"

# Chạy ứng dụng
run:
	@echo "🚀 Khởi động bot..."
	go run main.go

# Chạy tests
test:
	@echo "🧪 Chạy tests..."
	go test ./...

# Xóa files build
clean:
	@echo "🧹 Dọn dẹp files build..."
	rm -rf bin/
	go clean
	@echo "✅ Dọn dẹp hoàn thành"

# Kiểm tra cấu hình
check-config:
	@echo "🔍 Kiểm tra cấu hình..."
	@if [ ! -f .env ]; then \
		echo "❌ File .env không tồn tại. Chạy 'make setup' để tạo"; \
		exit 1; \
	fi
	@echo "✅ File .env tồn tại"
	@if grep -q "your_telegram_bot_token_here" .env; then \
		echo "⚠️  Vui lòng cập nhật TELEGRAM_BOT_TOKEN trong file .env"; \
	fi
	@if grep -q "your_chat_id_here" .env; then \
		echo "⚠️  Vui lòng cập nhật TELEGRAM_CHAT_ID trong file .env"; \
	fi

# Format code
format:
	@echo "🎨 Format code..."
	go fmt ./...
	@echo "✅ Code đã được format"

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run
	@echo "✅ Linting hoàn thành"

# Install development tools
dev-tools:
	@echo "🛠️  Cài đặt development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools đã được cài đặt"

# Show project info
info:
	@echo "📊 Thông tin dự án:"
	@echo "  Tên: Crypto Analysis Telegram Bot"
	@echo "  Ngôn ngữ: Go"
	@echo "  Version: 1.0.0"
	@echo "  Tác giả: Your Name"
	@echo ""
	@echo "📁 Cấu trúc dự án:"
	@tree -I 'node_modules|.git|bin|vendor' -a 