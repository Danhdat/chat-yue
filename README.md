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

## 🔄 Logic Chi tiết Function `FetchAndSaveAllSymbolsVolume()`

### 📊 Luồng xử lý tổng quan:
```go
func (s *AutoVolumeService) FetchAndSaveAllSymbolsVolume() error {
    // 1. Lấy danh sách symbols
    symbols = ["BTCUSDT", "ETHUSDT", ...]           // Regular symbols
    alphaSymbols = ["ALPHA_1", "ALPHA_2", ...]      // Alpha symbols  
    allSymbols = symbols + alphaSymbols
    
    // 2. Xử lý từng symbol
    for _, originalSymbol := range allSymbols {
        // 3. Phân loại và resolve symbol
        // 4. Gọi API tương ứng
        // 5. Xử lý dữ liệu klines
        // 6. Lưu vào database
    }
}
```

### 🎯 Chi tiết xử lý từng loại Symbol:
Regular Symbols: "BTCUSDT" → API: "BTCUSDT" → DB: "BTCUSDT"
Alpha Symbols: "ALPHA_1" → API: "ALPHA_1USDT" → DB: "Koma"

#### **Regular Symbols (ví dụ: "BTCUSDT")**
```go
originalSymbol = "BTCUSDT"
isAlphaSymbol("BTCUSDT") = false → symbolType = 0
resolvedSymbol = "BTCUSDT" (giữ nguyên)
API call: fetchRegularKlines("BTCUSDT")
Lưu DB với Symbol = "BTCUSDT", Type = 0
```

#### **Alpha Symbols (ví dụ: "ALPHA_1")**
```go
originalSymbol = "ALPHA_1"
isAlphaSymbol("ALPHA_1") = true → symbolType = 1
actualSymbol = "Koma" (từ GetNameByAlphaSymbol("ALPHA_1"))
resolvedSymbol = "Koma" (tên thật của token)
API call: fetchAlphaKlines("ALPHA_1USDT")
Lưu DB với Symbol = "Koma", Type = 1
```

### 🔧 Xử lý dữ liệu Klines:
```go
// 1. Lấy 23 nến từ API
klines = fetchKlines(symbol, limit=23)

// 2. Loại bỏ nến cuối (chưa đóng)
if len(klines) > 1 {
    klines = klines[:len(klines)-1]  // Bỏ nến cuối
}

// 3. Lấy 22 nến gần nhất
recentKlines = klines
if len(klines) > 22 {
    recentKlines = klines[len(klines)-22:]
}

// 4. Chuyển đổi thành records
for _, k := range recentKlines {
    record := models.AutoVolumeRecord{
        Symbol:           resolvedSymbol,  // Tên thật (BTCUSDT hoặc Koma)
        OpenTime:         parseKlineValue(k[0]),
        QuoteAssetVolume: parseFloat(k[7]),
        OpenPrice:        parseFloat(k[1]),
        ClosePrice:       parseFloat(k[4]),
        HighPrice:        parseFloat(k[2]),
        LowPrice:         parseFloat(k[3]),
        Type:             symbolType,      // 0: Regular, 1: Alpha
    }
    records = append(records, record)
}

// 5. Lưu vào database
volumeRepo.ReplaceAllForSymbol(resolvedSymbol, records)
```

### 🎯 Kết quả cuối cùng:
| Loại | Original Symbol | API Call | DB Symbol | Type |
|------|----------------|----------|-----------|------|
| **Regular** | `"BTCUSDT"` | `"BTCUSDT"` | `"BTCUSDT"` | `0` |
| **Alpha** | `"ALPHA_1"` | `"ALPHA_1USDT"` | `"Koma"` | `1` |

### 🔍 Điểm quan trọng:
- ✅ **API calls**: Sử dụng format gốc (`BTCUSDT`, `ALPHA_1USDT`)
- ✅ **Database storage**: Sử dụng tên thật (`BTCUSDT`, `Koma`)
- ✅ **Symbol resolution**: Alpha symbols được resolve thành tên dễ hiểu
- ✅ **Type tracking**: Phân biệt loại symbol qua field `Type`
- ✅ **Error handling**: Bỏ qua symbols không thể resolve
- ✅ **Rate limiting**: Sleep 100ms giữa các API calls

### 📝 Log output:
```
Đã cập nhật 22 records volume cho BTCUSDT (original: BTCUSDT)
Đã cập nhật 22 records volume cho Koma (original: ALPHA_1)
```

## 🔍 So sánh các Hàm Phân tích Mô hình Kỹ thuật

### 📊 Cấu trúc Dữ liệu
Tất cả các hàm phân tích đều nhận tham số `records []models.AutoVolumeRecord` với 22 nến đã đóng:

```go
// Cấu trúc dữ liệu:
records[0]  = nến 22 (mới nhất đã đóng) ← Phân tích chính
records[1]  = nến 21 (đã đóng trước đó) ← So sánh
records[2]  = nến 20 (đã đóng trước đó nữa)
...
records[21] = nến 1  (cũ nhất)
```

### 🎯 So sánh chi tiết các Hàm

| Hàm | Tham số | Logic Phân tích | Điều kiện Phát hiện | Ưu tiên |
|-----|---------|-----------------|-------------------|---------|
| **`detectBreakout`** | `records, averageCandlestickBody` | Phân tích trên `records[0]` (nến 22) | - Nến tăng (1)<br>- Thân dài > 1.5x trung bình<br>- Volume > 1.2x nến trước<br>- Phá vỡ resistance | 🥇 **1st** |
| **`detectEngulfing`** | `records` | So sánh `records[1]` vs `records[0]` | - Nến trước giảm (0) + Nến hiện tăng (1)<br>- Volume > 1.2x nến trước<br>- Giá engulfing | 🥈 **2nd** |
| **`detectHammer`** | `records` | Phân tích trên `records[0]` (nến 22) | - Thân ≤ 30% tổng dài<br>- Bóng dưới ≥ 2x thân<br>- Bóng trên ≤ 0.5x thân<br>- Sau downtrend | 🥉 **3rd** |
| **`detectDojiSpecial`** | `records` | Phân tích trên `records[0]` (nến 22) | - Thân ≤ 10% tổng dài<br>- Dragonfly/Gravestone patterns<br>- Kiểm tra xu hướng | 🏅 **4th** |

### 🔧 Chi tiết từng Hàm

#### 🚀 `detectBreakout` - Phá vỡ Kháng cự
```go
func detectBreakout(records []models.AutoVolumeRecord, averageCandlestickBody float64) PatternDetectionResult {
    // Phân tích trên nến 22 (records[0])
    record21 := records[0] // Nến 22 (đã đóng gần nhất)
    record20 := records[1] // Nến 21
    
    // Tính resistance từ nến 3-19
    resistance := calculateResistance(records)
    
    // Điều kiện: Nến tăng + Thân dài + Volume cao + Phá vỡ
    if record21.Candlestick() == 1 &&
       record21.IsCandlestickBodyLong(averageCandlestickBody, 1.5) &&
       record21.QuoteAssetVolume > record20.QuoteAssetVolume*1.2 &&
       record20.ClosePrice < resistance &&
       record21.ClosePrice > resistance {
        return PatternDetectionResult{IsDetected: true, Direction: 1}
    }
}
```

#### ⚙️ `detectEngulfing` - Mô hình Nuốt
```go
func detectEngulfing(records []models.AutoVolumeRecord) PatternDetectionResult {
    // So sánh nến 21 vs nến 22
    // records[1] = nến 21, records[0] = nến 22
    
    // Bullish Engulfing: Nến 21 giảm + Nến 22 tăng
    if records[1].Candlestick() == 0 && records[0].Candlestick() == 1 &&
       records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
       records[0].OpenPrice < records[1].ClosePrice &&
       records[0].ClosePrice > records[1].OpenPrice {
        return PatternDetectionResult{IsDetected: true, Direction: 1}
    }
    
    // Bearish Engulfing: Nến 21 tăng + Nến 22 giảm
    if records[1].Candlestick() == 1 && records[0].Candlestick() == 0 &&
       records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
       records[0].OpenPrice > records[1].ClosePrice &&
       records[0].ClosePrice < records[1].OpenPrice {
        return PatternDetectionResult{IsDetected: true, Direction: 2}
    }
}
```

#### 🔨 `detectHammer` - Mô hình Búa
```go
func detectHammer(records []models.AutoVolumeRecord) PatternDetectionResult {
    // Phân tích trên nến 22 (records[0])
    body := records[0].CandlestickBody()
    totalLength := records[0].CandlestickLength()
    upperShadow := records[0].CandlestickUpperShadow()
    lowerShadow := records[0].CandlestickLowerShadow()
    
    // Tiêu chuẩn Hammer chuyên nghiệp
    validBodySize := body <= totalLength*0.3      // Thân ≤ 30%
    validLowerShadow := lowerShadow >= body*2     // Bóng dưới ≥ 2x thân
    minimalUpperShadow := upperShadow <= body*0.5 // Bóng trên ≤ 0.5x thân
    shadowRatio := lowerShadow >= upperShadow*3   // Bóng dưới ≥ 3x bóng trên
    validPosition := isDowntrend                  // Sau downtrend
    
    if validBodySize && validLowerShadow && minimalUpperShadow && shadowRatio && validPosition {
        return PatternDetectionResult{IsDetected: true, Direction: 1}
    }
}
```

#### 🎭 `detectDojiSpecial` - Mô hình Doji
```go
func detectDojiSpecial(records []models.AutoVolumeRecord) PatternDetectionResult {
    // Phân tích trên nến 22 (records[0])
    body := records[0].CandlestickBody()
    totalLength := records[0].CandlestickLength()
    upperShadow := records[0].CandlestickUpperShadow()
    lowerShadow := records[0].CandlestickLowerShadow()
    
    // Kiểm tra điều kiện Doji
    if body > totalLength*0.1 { // Thân > 10% → không phải Doji
        return PatternDetectionResult{IsDetected: false}
    }
    
    // Dragonfly Doji: Bóng dưới dài, bóng trên ngắn
    if upperShadow <= totalLength*0.05 && lowerShadow >= totalLength*0.3 {
        return PatternDetectionResult{IsDetected: true, Direction: 1}
    }
    
    // Gravestone Doji: Bóng trên dài, bóng dưới ngắn
    if lowerShadow <= totalLength*0.05 && upperShadow >= totalLength*0.3 {
        return PatternDetectionResult{IsDetected: true, Direction: 2}
    }
}
```

### 🎯 Thứ tự Ưu tiên Phân tích
```go
// Ưu tiên: Breakout > Engulfing > Hammer > Doji
direction := 0
if breakoutResult.IsDetected {
    direction = breakoutResult.Direction
} else if engulfingResult.IsDetected {
    direction = engulfingResult.Direction
} else if hammerResult.IsDetected {
    direction = hammerResult.Direction
} else if dojiResult.IsDetected {
    direction = dojiResult.Direction
}
```

### 🔍 Đặc điểm chung
- ✅ **Tất cả đều phân tích trên nến đã đóng** (không phải nến đang hình thành)
- ✅ **Kiểm tra volume tăng** (1.2x so với nến trước)
- ✅ **Trả về `PatternDetectionResult`** với Pattern, Confirmation, Direction
- ✅ **Có kiểm tra điều kiện biên** để tránh lỗi
- ✅ **Hỗ trợ cả Bullish (1) và Bearish (2)** directions

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