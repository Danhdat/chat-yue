# Crypto Analysis Telegram Bot

Bot Telegram phân tích cryptocurrency với các chỉ báo kỹ thuật RSI, EMA, và MACD được viết bằng Golang.

## 🚀 Tính năng

- 📊 **Phân tích kỹ thuật** với các chỉ báo:
  - RSI (Relative Strength Index)
  - EMA (Exponential Moving Average) - 20, 50, 200 ngày
  - MACD (Moving Average Convergence Divergence)
- 💰 **Xem giá real-time** của các cryptocurrency
- 🏆 **Top cryptocurrencies** theo market cap
- ⚠️ **Đánh giá rủi ro** dựa trên các chỉ báo
- 🔧 **Hỗ trợ proxy** để bypass firewall

## 📋 Yêu cầu hệ thống

- Go 1.24.3 trở lên
- Telegram Bot Token
- Kết nối internet (có thể qua proxy)

## 🛠️ Cài đặt

### 1. Clone repository
```bash
git clone <repository-url>
cd BOT-BTC-TELEGRAM-SYSTEM
```

### 2. Cài đặt dependencies
```bash
go mod tidy
```

### 3. Cấu hình môi trường
Tạo file `.env` từ file mẫu:
```bash
cp config.env.example .env
```

Chỉnh sửa file `.env` với thông tin của bạn:
```env
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here

# Proxy Configuration (tùy chọn)
PROXY_ENABLED=true
PROXY_URL=http://proxy.example.com:8080
PROXY_USERNAME=proxy_username
PROXY_PASSWORD=proxy_password

# Server Configuration
SERVER_PORT=8080
LOG_LEVEL=info
```

### 4. Tạo Telegram Bot

1. Mở Telegram và tìm `@BotFather`
2. Gửi lệnh `/newbot`
3. Đặt tên cho bot
4. Đặt username cho bot (phải kết thúc bằng 'bot')
5. Lưu lại Bot Token

### 5. Lấy Chat ID (tùy chọn)

Nếu bạn muốn bot chỉ trả lời cho một chat cụ thể:
1. Gửi tin nhắn cho bot của bạn
2. Truy cập: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. Tìm `chat.id` trong response JSON
4. Thêm `TELEGRAM_CHAT_ID=your_chat_id` vào file `.env`

Nếu không cấu hình, bot sẽ trả lời tất cả users gửi tin nhắn.

## 🚀 Chạy ứng dụng

```bash
go run main.go
```

## 📱 Sử dụng Bot

### Các lệnh có sẵn:

- `/start` - Khởi động bot
- `/help` - Xem hướng dẫn
- `/analyze <symbol>` - Phân tích kỹ thuật (ví dụ: `/analyze BTCUSDT`)
- `/price <symbol>` - Xem giá (ví dụ: `/price ETHUSDT`)

### Ví dụ sử dụng:

```
/analyze BTCUSDT
📊 Phân tích kỹ thuật BTCUSDT

💰 Giá hiện tại: $43,250.50

📈 Chỉ báo kỹ thuật:
• RSI (14): 65.5
• EMA (20): $2,640.25
• EMA (50): $2,580.50
• EMA (200): $2,450.75
• MACD: 15.25
• MACD Signal: 12.50

🎯 Tín hiệu:
RSI: Trung tính (30-70)
EMA: Xu hướng tăng (Golden Cross)
MACD: Tín hiệu tăng

⚠️ Mức độ rủi ro: THẤP

🕐 Phân tích lúc: 14:30:25 15/12/2024

/price ETHUSDT
💰 Giá ETHUSDT

💵 Giá hiện tại: $2,650.75
📊 Thay đổi 24h: +2.5% (+$1,050.25)
💎 Market Cap: $850,000,000,000
📈 Volume 24h: $25,000,000,000
🕐 Cập nhật: 14:30:25
```

## 🔧 Cấu hình Proxy

Bot hỗ trợ proxy để bypass firewall:

```env
PROXY_ENABLED=true
PROXY_URL=http://proxy.example.com:8080
PROXY_USERNAME=your_username
PROXY_PASSWORD=your_password
```

## 📊 Chỉ báo kỹ thuật

### RSI (Relative Strength Index)
- **Quá mua**: > 70 (có thể giảm giá)
- **Quá bán**: < 30 (có thể tăng giá)
- **Trung tính**: 30-70

### EMA (Exponential Moving Average)
- **Golden Cross**: EMA20 > EMA50 > EMA200 (xu hướng tăng)
- **Death Cross**: EMA20 < EMA50 < EMA200 (xu hướng giảm)

### MACD (Moving Average Convergence Divergence)
- **Tín hiệu tăng**: MACD > Signal Line
- **Tín hiệu giảm**: MACD < Signal Line

## 🏗️ Cấu trúc dự án

```
BOT-BTC-TELEGRAM-SYSTEM/
├── config/
│   └── config.go          # Cấu hình ứng dụng
├── models/
│   └── crypto.go          # Models dữ liệu
├── services/
│   ├── crypto_api.go      # Service lấy dữ liệu crypto
│   ├── indicators.go      # Service tính toán chỉ báo
│   └── telegram_bot.go    # Service Telegram bot
├── main.go                # Entry point
├── go.mod                 # Dependencies
├── config.env.example     # File cấu hình mẫu
└── README.md              # Hướng dẫn sử dụng
```

## 🔒 Bảo mật

- Không commit file `.env` chứa thông tin nhạy cảm
- Sử dụng proxy nếu cần thiết
- Cập nhật dependencies thường xuyên

## 🐛 Xử lý lỗi

### Lỗi thường gặp:

1. **"TELEGRAM_BOT_TOKEN không được cấu hình"**
   - Kiểm tra file `.env` có chứa `TELEGRAM_BOT_TOKEN`

2. **"Lỗi cấu hình proxy"**
   - Kiểm tra URL proxy có đúng định dạng không
   - Kiểm tra thông tin xác thực proxy

3. **"Không đủ dữ liệu để phân tích"**
   - Một số coin mới có thể chưa đủ dữ liệu lịch sử

## 📈 Phát triển

### Thêm chỉ báo mới:

1. Thêm logic tính toán trong `services/indicators.go`
2. Cập nhật `models/crypto.go` để lưu trữ dữ liệu
3. Cập nhật `services/telegram_bot.go` để hiển thị

### Thêm API mới:

1. Tạo service mới trong thư mục `services/`
2. Cập nhật `config/config.go` nếu cần
3. Tích hợp vào bot

## 📄 License

MIT License

## 🤝 Đóng góp

Mọi đóng góp đều được chào đón! Vui lòng:

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📞 Hỗ trợ

Nếu gặp vấn đề, vui lòng tạo issue trên GitHub hoặc liên hệ qua Telegram. 