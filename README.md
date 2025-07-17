# BOT BTC TELEGRAM SYSTEM

Bot Telegram phân tích volume và kỹ thuật cho thị trường crypto, sử dụng Golang.

## 🚀 Tính năng nổi bật
- Tự động lấy dữ liệu volume 22 nến đã đóng gần nhất từ Binance (loại bỏ nến chưa đóng)
- Phân tích volume, phát hiện volume spike, cảnh báo tín hiệu mạnh/yếu
- Phát hiện mô hình nến đảo chiều (Bullish/Bearish Engulfing, Piercing)
- Gửi cảnh báo tự động qua Telegram
- Hỗ trợ các chỉ báo kỹ thuật khác (RSI, EMA, MACD...)

## 🏗️ Cấu trúc dự án
```
BOT BTC TELEGRAM SYSTEM/
├── config/                # Cấu hình ứng dụng
├── controllers/           # Xử lý request
├── models/                # Định nghĩa model, repository
├── routes/                # Định nghĩa route
├── services/              # Logic nghiệp vụ, phân tích, bot Telegram
├── utils/                 # Hàm tiện ích
├── main.go                # Entry point
├── go.mod, go.sum         # Quản lý dependencies
└── README.md              # Hướng dẫn sử dụng
```

## 📋 Yêu cầu hệ thống
- Go >= 1.18
- Telegram Bot Token
- Kết nối internet
- PostgreSQL (nếu dùng DB)

## ⚡ Cài đặt & Chạy bot
1. Clone repo:
   ```bash
git clone <repository-url>
cd BOT\ BTC\ TELEGRAM\ SYSTEM
```
2. Cài dependencies:
   ```bash
go mod tidy
```
3. Cấu hình file `.env` hoặc `config.go` (Telegram token, DB...)
4. Chạy bot:
   ```bash
go run main.go
```

## 🔎 Logic lấy và phân tích volume
- **Luôn lấy 23 nến gần nhất từ Binance**
- **Loại bỏ cây nến cuối cùng (nến chưa đóng)**
- **Chỉ phân tích trên 22 cây nến đã đóng**
- Khi phân tích volume:
  - Tính SMA 21 kỳ trên 21 nến đã đóng
  - So sánh volume nến mới nhất với SMA để phát hiện volume spike
  - Chỉ gửi cảnh báo khi volume đủ mạnh (theo ngưỡng cấu hình)

## 🛠️ Các lệnh Telegram hỗ trợ
- `/start` - Khởi động bot
- `/help` - Hướng dẫn sử dụng
- `/analyze <symbol>` - Phân tích kỹ thuật (ví dụ: `/analyze BTCUSDT`)
- `/price <symbol>` - Xem giá (ví dụ: `/price ETHUSDT`)

## 📝 Lưu ý kỹ thuật
- **Không sử dụng nến chưa đóng để phân tích volume**
- Khi lưu vào DB, chỉ lưu 22 nến đã đóng gần nhất
- Khi lấy dữ liệu từ DB để phân tích, lấy 23 nến gần nhất, loại bỏ nến cuối cùng nếu chưa đóng
- Đảm bảo đủ dữ liệu (ít nhất 22 nến đã đóng) để phân tích

## 🐞 Xử lý lỗi thường gặp
- Không đủ dữ liệu: Một số coin mới có thể chưa đủ 22 nến đã đóng
- Lỗi Telegram token: Kiểm tra lại cấu hình
- Lỗi kết nối Binance: Kiểm tra internet hoặc API limit

## 🤝 Đóng góp
- Fork repo, tạo branch mới, gửi pull request
- Đóng góp ý tưởng, báo lỗi qua GitHub Issue

## 📄 License
MIT License

---

Nếu cần hỗ trợ, hãy tạo issue hoặc liên hệ qua Telegram! 