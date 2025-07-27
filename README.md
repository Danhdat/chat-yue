# 🤖 BOT BTC TELEGRAM SYSTEM

Hệ thống bot Telegram tự động phân tích volume và phát hiện tín hiệu giao dịch cho các cặp tiền điện tử trên Binance.

## 🚀 Tính năng chính

### 📊 Phân tích Volume
- **Volume Spike Detection**: Phát hiện volume tăng đột biến (1.5x, 2x, 3x so với SMA21)
- **Real-time Monitoring**: Giám sát liên tục 22 cặp tiền chính và Alpha tokens
- **Multi-timeframe Analysis**: Phân tích dữ liệu 1 giờ với 23 nến gần nhất

### 🎯 Phát hiện Mô hình Kỹ thuật
- **🐂 Bullish Engulfing**: Mô hình đảo chiều tăng giá
- **🐻 Bearish Engulfing**: Mô hình đảo chiều giảm giá  
- **🚀 Breakout**: Phá vỡ kháng cự với volume cao
- **🔨 Hammer**: Mô hình búa đảo chiều
- **🎭 Doji Patterns**: Dragonfly Doji, Gravestone Doji

### 📱 Thông báo Telegram
- **Instant Alerts**: Gửi cảnh báo ngay lập tức khi phát hiện tín hiệu
- **Rich Formatting**: Thông báo với emoji và thông tin chi tiết
- **Channel Support**: Hỗ trợ gửi đến nhiều channel khác nhau
- **Daily/Weekly Stats**: Thống kê số lần xuất hiện tín hiệu

## 🏗️ Kiến trúc hệ thống

```
BOT BTC TELEGRAM SYSTEM/
├── main.go                 # Entry point
├── config/
│   └── config.go          # Cấu hình hệ thống
├── models/
│   ├── crypto.go          # Models cho crypto data
│   ├── database.go        # Database connection
│   ├── models.go          # Core models
│   ├── repository.go      # Data access layer
│   └── symbol.go          # Symbol management
├── services/
│   ├── analysis_service.go    # Phân tích kỹ thuật
│   ├── auto_volume_service.go # Auto volume monitoring
│   ├── crypto_api.go          # Binance API integration
│   ├── fetcher_service.go     # Data fetching
│   ├── indicators.go          # Technical indicators
│   ├── report_service.go      # Báo cáo
│   └── telegram_bot.go        # Telegram bot service
├── controllers/           # HTTP controllers
├── middleware/           # HTTP middleware
├── routes/              # API routes
└── utils/
    └── formatter.go     # Formatting utilities
```

## 🔧 Cài đặt và Chạy

### Yêu cầu hệ thống
- Go 1.19+
- MySQL/PostgreSQL
- Telegram Bot Token

### 1. Clone repository
```bash
git clone <repository-url>
cd BOT-BTC-TELEGRAM-SYSTEM
```

### 2. Cài đặt dependencies
```bash
go mod download
```

### 3. Cấu hình
Tạo file `.env` hoặc cấu hình trong `config/config.go`:
```env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=crypto_bot
DB_USER=root
DB_PASSWORD=password

TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHANNEL_ID=@your_channel

BINANCE_API_KEY=your_api_key
BINANCE_SECRET_KEY=your_secret_key
```

### 4. Chạy ứng dụng
```bash
go run main.go
```

## 📈 Cách hoạt động

### 1. Data Collection (Scheduler2)
- **Chạy mỗi giờ** (ví dụ: 9:00, 10:00, 11:00...)
- Lấy dữ liệu từ Binance API cho 22 cặp tiền chính
- Lấy dữ liệu từ Alpha API cho các token mới
- Lưu vào database với format chuẩn

### 2. Analysis & Notification (Scheduler3)  
- **Chạy mỗi giờ + 2 phút** (ví dụ: 9:02, 10:02, 11:02...)
- Phân tích volume và mô hình kỹ thuật
- Gửi cảnh báo qua Telegram khi phát hiện tín hiệu mạnh

### 3. Volume Analysis Logic
```go
// Phân tích volume theo thứ tự ưu tiên:
records22[0] = nến 22 (mới nhất) - Volume hiện tại
records22[1] = nến 21 
records22[2] = nến 20
...
records22[21] = nến 1 (cũ nhất)
```

## 🎯 Các loại Symbol

### Regular Symbols
- **Format**: `BTCUSDT`, `ETHUSDT`, `ADAUSDT`...
- **API**: Binance Spot API
- **Type**: 0

### Alpha Symbols  
- **Format**: `ALPHA_1`, `ALPHA_2`, `ALPHA_3`...
- **API**: Binance Alpha API (`ALPHA_1USDT`)
- **Resolved Name**: `Koma`, `Token2`, `Token3`...
- **Type**: 1

## 📊 Cấu trúc Database

### AutoVolumeRecord
```sql
CREATE TABLE auto_volume_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    symbol VARCHAR(50) NOT NULL,
    open_time BIGINT NOT NULL,
    quote_asset_volume DECIMAL(20,8) NOT NULL,
    open_price DECIMAL(20,8) NOT NULL,
    close_price DECIMAL(20,8) NOT NULL,
    high_price DECIMAL(20,8) NOT NULL,
    low_price DECIMAL(20,8) NOT NULL,
    type TINYINT DEFAULT 0, -- 0: Regular, 1: Alpha
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### NotificationLog
```sql
CREATE TABLE notification_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    symbol VARCHAR(50) NOT NULL,
    direction TINYINT DEFAULT 0, -- 0: Neutral, 1: Bullish, 2: Bearish
    type TINYINT DEFAULT 0, -- 0: Regular, 1: Alpha
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🔔 Format Thông báo Telegram

```
💰[CEX] [ALERT] Symbol: BTC
📅 Time: 2024-01-15 10:02:00
🚀 Volume: 1,234,567.89 (SMA21: 456,789.12)
💵 Price: $45,678.90
🎯 Strength: STRONG
🔥 Signal: 🚀 HIGH VOLUME SPIKE
🔖 Daily Occurrences: 3
✨ Pattern: ⚙️ Mô hình 🐂 Bullish Engulfing
📊 Confirmation: ✅ Tín hiệu đảo chiều tăng giá
💎 Weekly Occurrences: 12
```

## 🛠️ API Endpoints

### Volume Analysis
- `GET /api/volume/analysis` - Phân tích volume hiện tại
- `GET /api/volume/history` - Lịch sử volume
- `POST /api/volume/analyze` - Phân tích thủ công

### Symbols Management  
- `GET /api/symbols` - Danh sách symbols
- `POST /api/symbols` - Thêm symbol mới
- `DELETE /api/symbols/{id}` - Xóa symbol

### Notifications
- `GET /api/notifications` - Lịch sử thông báo
- `POST /api/notifications/test` - Test gửi thông báo

## 📝 Logs và Monitoring

### Log Levels
- **INFO**: Thông tin hoạt động bình thường
- **WARNING**: Cảnh báo về lỗi API hoặc dữ liệu
- **ERROR**: Lỗi nghiêm trọng cần xử lý

### Monitoring Metrics
- Số lượng symbols được xử lý
- Tỷ lệ thành công của API calls
- Số lượng thông báo đã gửi
- Thời gian phản hồi của hệ thống

## 🔒 Bảo mật

- **API Keys**: Lưu trữ an toàn trong environment variables
- **Rate Limiting**: Giới hạn số lượng API calls
- **Error Handling**: Xử lý lỗi gracefully
- **Data Validation**: Kiểm tra dữ liệu đầu vào

## 🤝 Đóng góp

1. Fork repository
2. Tạo feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Tạo Pull Request

## 📄 License

MIT License - xem file [LICENSE](LICENSE) để biết thêm chi tiết.

## 📞 Liên hệ

- **Email**: your-email@example.com
- **Telegram**: @your_username
- **GitHub**: [your-github](https://github.com/your-username)

---

⭐ **Star repository này nếu bạn thấy hữu ích!** 