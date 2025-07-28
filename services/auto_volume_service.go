package services

import (
	"chatbtc/models"
	"chatbtc/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type AutoVolumeService struct {
	volumeRepo          *models.AutoVolumeRecordRepository
	symbolRepo          *models.SymbolRepository
	notificationLogRepo *models.NotificationLogRepository
	alphaRepo           *models.AlphaSymbolRepository
	telegramBotService  *TelegramBotService
}

// Truyền TelegramBotService vào khi khởi tạo
func NewAutoVolumeService(telegramBotService *TelegramBotService) *AutoVolumeService {
	return &AutoVolumeService{
		volumeRepo:          models.NewAutoVolumeRecordRepository(),
		symbolRepo:          models.NewSymbolRepository(),
		notificationLogRepo: models.NewNotificationLogRepository(),
		alphaRepo:           models.NewAlphaSymbolRepository(),
		telegramBotService:  telegramBotService,
	}
}

// Hàm kiểm tra alpha symbol
func isAlphaSymbol(symbol string) bool {
	return strings.HasPrefix(symbol, "ALPHA_")
}

func (s *AutoVolumeService) fetchRegularKlines(symbol string) ([][]interface{}, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1h&limit=23", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, fmt.Errorf("lỗi parse regular klines: %w", err)
	}

	return klines, nil
}

func (s *AutoVolumeService) fetchAlphaKlines(symbol string) ([][]interface{}, error) {
	url := fmt.Sprintf("https://www.binance.com/bapi/defi/v1/public/alpha-trade/klines?interval=1h&limit=23&symbol=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cấu trúc response mới của API Alpha
	var alphaResponse struct {
		Code    string          `json:"code"`
		Message interface{}     `json:"message"`
		Data    [][]interface{} `json:"data"`
		Success bool            `json:"success"`
	}

	if err := json.Unmarshal(body, &alphaResponse); err != nil {
		return nil, fmt.Errorf("lỗi parse alpha klines: %w", err)
	}

	if !alphaResponse.Success || alphaResponse.Code != "000000" {
		return nil, fmt.Errorf("API Alpha trả về lỗi: %v", alphaResponse.Message)
	}

	return alphaResponse.Data, nil
}

func (s *AutoVolumeService) FetchAndSaveAllSymbolsVolume() error {
	symbols, err := s.symbolRepo.GetAllSymbols()
	if err != nil {
		return err
	}

	alphaSymbols, err := s.alphaRepo.GetAllAlphaSymbols()
	if err != nil {
		return fmt.Errorf("lỗi lấy alpha symbols: %w", err)
	}

	allSymbols := append(symbols, alphaSymbols...)

	for _, originalSymbol := range allSymbols {
		// Xác định symbolType và resolved symbol trước khi xử lý klines
		symbolType := 0
		resolvedSymbol := originalSymbol

		if isAlphaSymbol(originalSymbol) {
			symbolType = 1
			actualSymbol, err := s.alphaRepo.GetNameByAlphaSymbol(originalSymbol)
			if err != nil {
				// RẤT QUAN TRỌNG: Nếu không thể resolve alpha symbol, ta BỎ QUA symbol này.
				// Điều này đảm bảo chúng ta không lưu dữ liệu với một symbol không chính xác hoặc không xác định.
				fmt.Printf("Lỗi lấy tên symbol cho alpha symbol '%s': %v. Bỏ qua symbol này.\n", originalSymbol, err)
				continue
			}
			resolvedSymbol = actualSymbol // Gán lại resolvedSymbol nếu tìm thấy
		}

		// Lấy dữ liệu kline (dùng originalSymbol cho API call)
		var klines [][]interface{}
		var err error

		if symbolType == 1 {
			klines, err = s.fetchAlphaKlines(originalSymbol + "USDT")
		} else {
			klines, err = s.fetchRegularKlines(originalSymbol)
		}

		if err != nil {
			fmt.Printf("Lỗi lấy dữ liệu %s: %v\n", originalSymbol, err)
			continue
		}

		// Xử lý klines - ĐẢM BẢO CÓ ĐỦ 22 NẾN ĐÃ ĐÓNG
		// API trả về 23 nến: [nến 1, nến 2, ..., nến 22, nến 23]
		// Trong đó nến 23 là nến đang hình thành (chưa đóng)

		// Loại bỏ nến cuối cùng (nến 23 - đang hình thành) để chỉ lấy nến đã đóng
		if len(klines) > 1 {
			klines = klines[:len(klines)-1] // Kết quả: [nến 1, nến 2, ..., nến 22]
		}

		// Lấy 22 nến đã đóng (từ nến 1 đến nến 22)
		recentKlines := klines
		if len(klines) > 22 {
			recentKlines = klines[len(klines)-22:] // Lấy 22 nến cuối cùng
		}

		// Debug: In ra số lượng nến để kiểm tra
		fmt.Printf("Symbol %s: API trả về %d nến, sau xử lý có %d nến\n",
			resolvedSymbol, len(klines), len(recentKlines))

		loc := time.FixedZone("UTC+7", 7*60*60)
		var records []models.AutoVolumeRecord

		for _, k := range recentKlines {
			openTime := parseKlineValue(k[0])
			quoteAssetVolumeStr := k[7].(string)
			quoteAssetVolume, _ := strconv.ParseFloat(quoteAssetVolumeStr, 64)
			openPriceStr := k[1].(string)
			openPrice, _ := strconv.ParseFloat(openPriceStr, 64)
			closePriceStr := k[4].(string)
			closePrice, _ := strconv.ParseFloat(closePriceStr, 64)
			highPriceStr := k[2].(string)
			highPrice, _ := strconv.ParseFloat(highPriceStr, 64)
			lowPriceStr := k[3].(string)
			lowPrice, _ := strconv.ParseFloat(lowPriceStr, 64)

			record := models.AutoVolumeRecord{
				Symbol:           resolvedSymbol, // Dùng symbol đã resolve
				OpenTime:         openTime,
				QuoteAssetVolume: quoteAssetVolume,
				OpenPrice:        openPrice,
				ClosePrice:       closePrice,
				HighPrice:        highPrice,
				LowPrice:         lowPrice,
				CreatedAt:        time.Now().In(loc),
				UpdatedAt:        time.Now().In(loc),
				Type:             symbolType,
			}
			records = append(records, record)
		}

		// Lưu vào database với resolvedSymbol
		if err := s.volumeRepo.ReplaceAllForSymbol(resolvedSymbol, records); err != nil {
			fmt.Printf("Lỗi lưu DB %s: %v\n", resolvedSymbol, err)
		} else {
			fmt.Printf("Đã cập nhật %d records volume cho %s (original: %s)\n",
				len(records), resolvedSymbol, originalSymbol)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (s *AutoVolumeService) AnalyzeAndNotifyVolumes(channelID string) error {
	// Lấy tất cả symbols thay vì tất cả records
	reSymbols, err := s.symbolRepo.GetAllSymbols()
	if err != nil {
		return err
	}
	alphaSymbols, err := s.alphaRepo.GetAllAlphaName()
	if err != nil {
		return err
	}
	symbols := append(reSymbols, alphaSymbols...)

	log.Println("Analyzing volumes for ", len(symbols), "symbols")
	taService := NewTechnicalAnalysisService()

	// Map để theo dõi symbols đã xử lý để tránh trùng lặp
	processedSymbols := make(map[string]bool)
	loc := time.FixedZone("UTC+7", 7*60*60)

	for _, symbol := range symbols {
		// Kiểm tra nếu symbol đã được xử lý
		if processedSymbols[symbol] {
			continue
		}
		records22, _ := s.volumeRepo.GetLastNBySymbol(symbol, 22) // Lấy 22 nến đã đóng
		// Kiểm tra nếu không có dữ liệu
		if len(records22) == 0 {
			continue
		}

		var volumes []float64
		for _, r := range records22 {
			volumes = append(volumes, r.QuoteAssetVolume)
		}
		//chỉ tới cây nến 21
		var totalCandlestickLength float64 = 0
		var totalCandlestickBody float64 = 0
		for _, r := range records22[1:] { // Bỏ qua records22[0] (nến 22 - mới nhất đã đóng)
			totalCandlestickLength += r.CandlestickLength()
			totalCandlestickBody += r.CandlestickBody()
		}
		averageCandlestickBody := totalCandlestickBody / float64(len(records22)-1)

		volumeAnalysis := taService.analyzeVolumeFromFloat64(volumes)
		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			// Lấy bản ghi MỚI NHẤT (records22[0]) - nến 22 (đã đóng gần nhất)
			latestRecord := records22[0]

			// Lấy time hiện tại
			currentTime := time.Now().In(loc)
			formattedTime := currentTime.Format("2006-01-02 15:04:05")

			// Phân tích mô hình trên nến ĐÃ ĐÓNG (records22[1] và records22[2])
			breakoutResult := detectBreakout(records22, averageCandlestickBody)
			confirmation3 := breakoutResult.Confirmation
			pattern3 := breakoutResult.Pattern
			engulfingResult := detectEngulfing(records22)
			confirmation1 := engulfingResult.Confirmation
			pattern1 := engulfingResult.Pattern
			dojiResult := detectDojiSpecial(records22)
			confirmation2 := dojiResult.Confirmation
			pattern2 := dojiResult.Pattern
			hammerResult := detectHammer(records22)
			confirmation4 := hammerResult.Confirmation
			pattern4 := hammerResult.Pattern

			patternString := utils.FormatElements(pattern1, pattern2, pattern3, pattern4)
			confirmationString := utils.FormatElements(confirmation1, confirmation2, confirmation3, confirmation4)
			count, _ := s.notificationLogRepo.CountBySymbolToday(symbol)
			countofWeek, _ := s.notificationLogRepo.CountBySymbolThisWeek(symbol)

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
			alertHeader := "*[CEX]*"
			if latestRecord.Type == 1 {
				alertHeader = "*[ALPHA BINANCE]* 🚨" // Thêm icon cảnh báo đặc biệt
			}
			// Tạo chuỗi hiển thị các nến từ records22[4] đến records22[0]
			var candlestickPattern strings.Builder
			candlestickPattern.WriteString("💡 ")
			for i := 4; i >= 0; i-- {
				if records22[i].Candlestick() == 1 {
					candlestickPattern.WriteString("🟢")
				} else {
					candlestickPattern.WriteString("🔴")
				}
			}

			message := fmt.Sprintf("💰%s *[ALERT]* Symbol: *%s*\n"+
				"📅 Time: %s\n"+
				"🚀 Volume: *%s* (SMA21: %s)\n"+
				"💵 Price: *%s*\n"+
				"🎯 Strength: *%s*\n"+
				"🔥 Signal: *%s*\n"+
				"🔖 Daily Occurrences: %d\n"+
				"✨ Pattern: %s\n"+
				"📊 Confirmation: %s\n"+
				"💎 Weekly Occurrences: %d\n"+
				"%s\n",
				alertHeader,
				strings.TrimSuffix(latestRecord.Symbol, "USDT"),
				formattedTime,
				utils.FormatVolume(decimal.NewFromFloat(latestRecord.QuoteAssetVolume)),
				utils.FormatVolume(volumeAnalysis.VolumeSMA21),
				utils.FormatPrice(decimal.NewFromFloat(latestRecord.ClosePrice)),
				volumeAnalysis.VolumeStrength,
				volumeAnalysis.VolumeSignal,
				count+1,
				patternString,
				confirmationString,
				countofWeek+1,
				candlestickPattern.String(),
			)
			s.telegramBotService.SendTelegramToChannel(channelID, message)

			// Lưu log sau khi gửi - Sử dụng async để tối ưu performance
			notificationLog := &models.NotificationLog{
				Symbol:    symbol,
				CreatedAt: time.Now(),
				Direction: direction,
				Type:      latestRecord.Type,
			}
			// Sử dụng async insert để không block main thread
			s.notificationLogRepo.CreateAsync(notificationLog)
		}

		// Đánh dấu symbol đã được xử lý
		processedSymbols[symbol] = true
		time.Sleep(1 * time.Second)
	}
	time.Sleep(1 * time.Second)

	return nil
}

// Hàm phân tích volume cho 1 giá trị float64 (tương thích với analyzeVolume)
func (s *TechnicalAnalysisService) analyzeVolumeFromFloat64(volumes []float64) models.VolumeAnalysis {
	// ĐẢO NGƯỢC SLICE Ở ĐÂY nếu cần
	for i, j := 0, len(volumes)-1; i < j; i, j = i+1, j-1 {
		volumes[i], volumes[j] = volumes[j], volumes[i]
	}
	if len(volumes) < models.VOLUME_SMA_PERIOD+1 {
		return models.VolumeAnalysis{}
	}
	// Chuyển sang decimal.Decimal để dùng lại logic cũ nếu cần
	currentVolume := decimal.NewFromFloat(volumes[len(volumes)-1])
	var sum float64
	for i := len(volumes) - models.VOLUME_SMA_PERIOD; i < len(volumes); i++ {
		sum += volumes[i]
	}
	volumeSMA := sum / float64(models.VOLUME_SMA_PERIOD)
	log.Println("volumes:", volumes)
	log.Println("SUM:", sum)
	log.Println("volumeSMA:", volumeSMA)
	var volumeSignal, volumeStrength, confirmation string
	confirmation = "null"
	var volumeRatio decimal.Decimal
	if volumeSMA > 0 {
		volumeRatio = currentVolume.Div(decimal.NewFromFloat(volumeSMA))
	} else {
		volumeRatio = decimal.Zero
	}
	if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_3X)) {
		volumeSignal = "🔥 VOLUME EXPLOSION"
		volumeStrength = "EXTREME"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_2X)) {
		volumeSignal = "🚀 HIGH VOLUME SPIKE"
		volumeStrength = "STRONG"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_1_5X)) {
		volumeSignal = "📈 ABOVE AVERAGE VOLUME"
		volumeStrength = "MODERATE"
		confirmation = "Tín hiệu TRUNG BÌNH - Có sự quan tâm tăng lên"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(1.0)) {
		volumeSignal = "🟡 NORMAL VOLUME"
		volumeStrength = "NORMAL"
	} else {
		volumeSignal = "📉 LOW VOLUME"
		volumeStrength = "WEAK"
	}
	return models.VolumeAnalysis{
		CurrentVolume:  currentVolume,
		VolumeSMA21:    decimal.NewFromFloat(volumeSMA),
		VolumeRatio:    volumeRatio,
		VolumeSignal:   volumeSignal,
		VolumeStrength: volumeStrength,
		Confirmation:   confirmation,
	}
}

type PatternDetectionResult struct {
	Pattern      string
	Confirmation string
	IsDetected   bool
	Direction    int
}

func detectEngulfing(records []models.AutoVolumeRecord) PatternDetectionResult {
	// records[1] = nến 21
	// records[0] = nến 22 (đã đóng gần nhất)

	if records[1].Candlestick() == 0 &&
		records[0].Candlestick() == 1 &&
		records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
		records[0].OpenPrice < records[1].ClosePrice &&
		records[0].ClosePrice > records[1].OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐂 Bullish Engulfing",
			Confirmation: "✅ Đây là một tín hiệu đảo chiều tăng giá rất mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng giảm. Nó cho thấy phe mua đã hoàn toàn áp đảo phe bán",
			IsDetected:   true,
			Direction:    1,
		}
	} else if records[1].Candlestick() == 1 &&
		records[0].Candlestick() == 0 &&
		records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
		records[0].OpenPrice > records[1].ClosePrice &&
		records[0].ClosePrice < records[1].OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐻 Bearish Engulfing",
			Confirmation: "🍎 Đây là một tín hiệu đảo chiều giảm giá mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng tăng. Nó cho thấy phe bán đã hoàn toàn áp đảo phe mua",
			IsDetected:   true,
			Direction:    2,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

func detectBreakout(records []models.AutoVolumeRecord, averageCandlestickBody float64) PatternDetectionResult {
	if len(records) < 8 { // Cần ít nhất 8 nến để có nến 15-19
		return PatternDetectionResult{IsDetected: false}
	}

	// QUAN TRỌNG: Phân tích Breakout trên nến ĐÃ ĐÓNG
	// records[0] = nến 22 (đã đóng gần nhất)
	// records[1] = nến 21 (đã đóng trước đó)
	record20 := records[1] // Nến 21
	record21 := records[0] // Nến 22 (đã đóng gần nhất)

	// Tính resistance level (cao nhất của 5 nến trước nến hiện tại)
	resistance := calculateResistance(records)
	log.Println("resistance:", resistance, "symbols", record21.Symbol)

	if record21.Candlestick() == 1 &&
		record21.IsCandlestickBodyLong(averageCandlestickBody, 1.5) &&
		record21.QuoteAssetVolume > record20.QuoteAssetVolume*1.2 &&
		record20.ClosePrice < resistance && // Nến trước chưa phá vỡ
		record21.ClosePrice > resistance { // Nến hiện tại phá vỡ
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐂 Breakout",
			Confirmation: "✅ Tín hiệu breakout: Giá đóng cửa vượt qua resistance",
			IsDetected:   true,
			Direction:    1,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

// Tính resistance level (cao nhất của 16 nến trước nến hiện tại)
func calculateResistance(records []models.AutoVolumeRecord) float64 {
	// Kiểm tra điều kiện biên
	if len(records) < 20 { // Cần ít nhất từ records[1] đến records[19]
		return 0
	}
	// Xác định phạm vi nến 3-19 (tương ứng records[19] đến records[3])
	// Vì:
	// records[0] = nến 22 (mới nhất)
	// CORRECTED RANGE: Nến 3-19 tương ứng với records[19] đến records[3]
	startIdx := 19 // nến 3
	endIdx := 3    // nến 19
	if startIdx >= len(records) || endIdx >= len(records) {
		return 0
	}

	resistance := records[startIdx].HighPrice
	for i := startIdx; i >= endIdx; i-- { // Lặp từ nến 3 đến 19
		if records[i].HighPrice > resistance {
			resistance = records[i].HighPrice
		}
	}

	return resistance
}

func detectHammer(records []models.AutoVolumeRecord) PatternDetectionResult {
	// Kiểm tra điều kiện biên
	if len(records) < 7 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// QUAN TRỌNG: Phân tích Hammer trên nến ĐÃ ĐÓNG (records[0]) - nến 22 (mới nhất đã đóng)
	// records[0] = nến 22 (đã đóng gần nhất) - PHÂN TÍCH Ở ĐÂY
	// records[1] = nến 21 (đã đóng trước đó)
	isDowntrend := checkDowntrendFromIndex(records, 1, 5)
	body := records[0].CandlestickBody()
	totalLength := records[0].CandlestickLength()
	upperShadow := records[0].CandlestickUpperShadow()
	lowerShadow := records[0].CandlestickLowerShadow()

	// Tiêu chuẩn nhận diện Hammer chuyên nghiệp
	validBodySize := body <= totalLength*0.3      // Thân ≤ 30% tổng chiều dài
	validLowerShadow := lowerShadow >= body*2     // Bóng dưới ≥ 2x thân
	minimalUpperShadow := upperShadow <= body*0.5 // Bóng trên ≤ 0.5x thân
	shadowRatio := lowerShadow >= upperShadow*3   // Bóng dưới dài gấp 3x bóng trên
	validPosition := isDowntrend                  // Xuất hiện sau downtrend

	if validBodySize && validLowerShadow && minimalUpperShadow && shadowRatio && validPosition {
		// Phân loại Hammer
		var direction int
		hammerType := "🐂 Bullish"
		confidence := "Tín hiệu mạnh"
		if records[0].ClosePrice < records[0].OpenPrice {
			hammerType = "🐻 Bearish (Hanging Man)"
			confidence = "Cần nến tăng xác nhận"
			direction = 2
		} else {
			direction = 1
		}

		return PatternDetectionResult{
			Pattern: fmt.Sprintf("⚙️ Mô hình Hammer (%s)", hammerType),
			Confirmation: fmt.Sprintf("✅ %s - Thân: %.2f%%, Bóng dưới: %.2f%%, Bóng trên: %.2f%%",
				confidence,
				(body/totalLength)*100,
				(lowerShadow/totalLength)*100,
				(upperShadow/totalLength)*100),
			IsDetected: true,
			Direction:  direction,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

func detectDojiSpecial(records []models.AutoVolumeRecord) PatternDetectionResult {
	const (
		bodyThreshold   = 0.1  // Thân nến ≤ 10% tổng độ dài
		shadowThreshold = 0.05 // Bóng ngắn ≤ 5% tổng độ dài
		minShadowRatio  = 2.0  // Bóng dài phải gấp ít nhất 2 lần thân
	)

	// Kiểm tra điều kiện biên - cần ít nhất 7 nến để phân tích
	if len(records) < 7 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// QUAN TRỌNG: Phân tích Doji trên nến ĐÃ ĐÓNG (records[0]) - nến 22 (mới nhất đã đóng)
	// records[0] = nến 22 (đã đóng gần nhất) - PHÂN TÍCH Ở ĐÂY
	// records[1] = nến 21 (đã đóng trước đó)
	candle := records[0] // Nến 22 (đã đóng gần nhất)
	body := candle.CandlestickBody()
	totalLength := candle.CandlestickLength()
	upperShadow := candle.CandlestickUpperShadow()
	lowerShadow := candle.CandlestickLowerShadow()

	// KIỂM TRA ĐIỀU KIỆN ĐỂ TRÁNH NaN
	// Nếu totalLength = 0 (HighPrice = LowPrice), không thể phân tích
	if totalLength <= 0 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Nếu body = 0, không thể tính tỷ lệ
	if body <= 0 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Bỏ qua nếu không phải Doji (thân quá lớn)
	if body > totalLength*bodyThreshold {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Kiểm tra xu hướng trước đó (5 nến trước records[0])
	// Sử dụng records[1] đến records[5] để kiểm tra xu hướng
	isDowntrend := checkDowntrendFromIndex(records, 1, 5)
	isUptrend := checkUptrendFromIndex(records, 1, 5)

	// Kiểm tra Dragonfly Doji (bóng dưới dài, bóng trên rất ngắn)
	if upperShadow <= totalLength*shadowThreshold &&
		lowerShadow >= totalLength*0.3 && // Bóng dưới ≥ 30% tổng độ dài
		lowerShadow >= body*minShadowRatio {

		// Dragonfly Doji chỉ có ý nghĩa sau downtrend
		if isDowntrend {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐂 Dragonfly Doji",
				Confirmation: "✅ Tín hiệu tăng mạnh sau downtrend - Phe mua đang tích lũy",
				IsDetected:   true,
				Direction:    1,
			}
		} else {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐂 Dragonfly Doji (Yếu)",
				Confirmation: "⚠️ Dragonfly Doji xuất hiện không sau downtrend - Tín hiệu yếu hơn",
				IsDetected:   true,
				Direction:    1,
			}
		}
	}

	// Kiểm tra Gravestone Doji (bóng trên dài, bóng dưới rất ngắn)
	if lowerShadow <= totalLength*shadowThreshold &&
		upperShadow >= totalLength*0.3 && // Bóng trên ≥ 30% tổng độ dài
		upperShadow >= body*minShadowRatio {

		// Gravestone Doji chỉ có ý nghĩa sau uptrend
		if isUptrend {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐻 Gravestone Doji",
				Confirmation: "🍎 Tín hiệu giảm mạnh sau uptrend - Phe bán đang áp đảo",
				IsDetected:   true,
				Direction:    2,
			}
		} else {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐻 Gravestone Doji (Yếu)",
				Confirmation: "⚠️ Gravestone Doji xuất hiện không sau uptrend - Tín hiệu yếu hơn",
				IsDetected:   true,
				Direction:    2,
			}
		}
	}

	// Kiểm tra Four Price Doji (thân rất nhỏ, bóng trên và dưới đều ngắn)
	if body <= totalLength*0.05 && // Thân ≤ 5% tổng độ dài
		upperShadow <= totalLength*0.1 && // Bóng trên ≤ 10%
		lowerShadow <= totalLength*0.1 { // Bóng dưới ≤ 10%

		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🔄 Four Price Doji",
			Confirmation: "🔄 Tín hiệu indecision - Thị trường đang cân bằng, chờ breakout",
			IsDetected:   true,
			Direction:    0, // Không xác định hướng
		}
	}

	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

// Hàm kiểm tra downtrend từ index cụ thể
func checkDowntrendFromIndex(records []models.AutoVolumeRecord, startIdx, endIdx int) bool {
	// Kiểm tra điều kiện biên
	if startIdx >= len(records) || endIdx >= len(records) || startIdx > endIdx {
		return false
	}

	// Tính số lượng nến giảm trong khoảng từ startIdx đến endIdx
	downCount := 0
	totalCandles := endIdx - startIdx + 1

	for i := startIdx; i <= endIdx; i++ {
		if records[i].Candlestick() == 0 { // Nến giảm
			downCount++
		}
	}

	// Xác định xu hướng giảm (ít nhất 60% nến là giảm)
	return float64(downCount)/float64(totalCandles) >= 0.6
}

// Hàm kiểm tra uptrend từ index cụ thể
func checkUptrendFromIndex(records []models.AutoVolumeRecord, startIdx, endIdx int) bool {
	// Kiểm tra điều kiện biên
	if startIdx >= len(records) || endIdx >= len(records) || startIdx > endIdx {
		return false
	}

	// Tính số lượng nến tăng trong khoảng từ startIdx đến endIdx
	upCount := 0
	totalCandles := endIdx - startIdx + 1

	for i := startIdx; i <= endIdx; i++ {
		if records[i].Candlestick() == 1 { // Nến tăng
			upCount++
		}
	}

	// Xác định xu hướng tăng (ít nhất 60% nến là tăng)
	return float64(upCount)/float64(totalCandles) >= 0.6
}

type Scheduler2 struct {
	autoVolumeService *AutoVolumeService
	stopChan          chan bool
}

// Truyền channelID vào khi khởi tạo Scheduler2
func NewScheduler2(autoVolumeService *AutoVolumeService) *Scheduler2 {
	return &Scheduler2{
		autoVolumeService: autoVolumeService,
		stopChan:          make(chan bool),
	}
}

func (s *Scheduler2) Start() {
	log.Println("Scheduler Volume started")
	// Hàm helper để tính thời gian đến giờ tiếp theo
	nextHour := func() time.Time {
		now := time.Now()
		next := now.Truncate(time.Hour).Add(time.Hour + 2*time.Minute)
		return next
	}
	// Tạo timer với thời gian đến giờ tiếp theo
	timer := time.NewTimer(time.Until(nextHour()))
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			go s.Run()
			// Đặt lại timer cho giờ tiếp theo
			timer.Reset(time.Until(nextHour()))
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler2) Stop() {
	s.stopChan <- true
}

func (s *Scheduler2) Run() {
	log.Println("Running update")
	if err := s.autoVolumeService.FetchAndSaveAllSymbolsVolume(); err != nil {
		log.Printf("Lỗi khi cập nhật dữ liệu: %v", err)
	}
	log.Println("Update completed")

}

type Scheduler3 struct {
	autoVolumeService *AutoVolumeService
	channelID         string
	stopChan          chan bool
}

func NewScheduler3(autoVolumeService *AutoVolumeService, channelID string) *Scheduler3 {
	return &Scheduler3{
		autoVolumeService: autoVolumeService,
		channelID:         channelID,
		stopChan:          make(chan bool),
	}
}

func (s *Scheduler3) Start() {
	// Hàm helper để tính thời gian đến giờ:02 phút tiếp theo
	nextSchedule := func() time.Time {
		now := time.Now()
		// Cắt lẻ đến giờ, sau đó thêm 1 giờ + 4 phút (ví dụ: 8:30 → 9:02:00)
		next := now.Truncate(time.Hour).Add(time.Hour + 4*time.Minute)
		return next
	}
	// Tạo timer với thời gian đến lần chạy tiếp theo (9:02:00 nếu now là 8:30:00)
	timer := time.NewTimer(time.Until(nextSchedule()))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			go s.Run()
			timer.Reset(time.Until(nextSchedule()))
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler3) Run() {
	if err := s.autoVolumeService.AnalyzeAndNotifyVolumes(s.channelID); err != nil {
		log.Printf("Lỗi khi phân tích và gửi cảnh báo: %v", err)
	}
	log.Println("Analyze and notify completed")
}
func (s *Scheduler3) Stop() {
	s.stopChan <- true
}

func parseKlineValue(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
