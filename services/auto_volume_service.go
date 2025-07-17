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
	volumeRepo         *models.AutoVolumeRecordRepository
	symbolRepo         *models.SymbolRepository
	telegramBotService *TelegramBotService
}

// Truyền TelegramBotService vào khi khởi tạo
func NewAutoVolumeService(telegramBotService *TelegramBotService) *AutoVolumeService {
	return &AutoVolumeService{
		volumeRepo:         models.NewAutoVolumeRecordRepository(),
		symbolRepo:         models.NewSymbolRepository(),
		telegramBotService: telegramBotService,
	}
}

func (s *AutoVolumeService) FetchAndSaveAllSymbolsVolume() error {
	symbols, err := s.symbolRepo.GetAllSymbols()
	if err != nil {
		return err
	}
	for _, symbol := range symbols {
		// Lấy dữ liệu kline
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1h&limit=23", symbol)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Lỗi lấy dữ liệu %s: %v\n", symbol, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var klines [][]interface{}
		if err := json.Unmarshal(body, &klines); err != nil || len(klines) == 0 {
			fmt.Printf("Lỗi parse dữ liệu %s: %v\n", symbol, err)
			continue
		}
		// Loại bỏ cây nến cuối cùng (chưa đóng) nếu có nhiều hơn 1 nến
		if len(klines) > 1 {
			klines = klines[:len(klines)-1]
		}
		// Lấy 22 nến đã đóng gần nhất
		recentKlines := klines
		if len(klines) > 22 {
			recentKlines = klines[len(klines)-22:]
		}

		loc := time.FixedZone("UTC+7", 7*60*60)

		// Tạo slice để lưu tất cả records cho symbol này
		var records []models.AutoVolumeRecord

		for _, k := range recentKlines {
			openTime := k[0].(float64)
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
				Symbol:           symbol,
				OpenTime:         openTime,
				QuoteAssetVolume: quoteAssetVolume,
				OpenPrice:        openPrice,
				ClosePrice:       closePrice,
				HighPrice:        highPrice,
				LowPrice:         lowPrice,
				CreatedAt:        time.Now().In(loc),
				UpdatedAt:        time.Now().In(loc),
			}
			records = append(records, record)
		}

		// Thay thế tất cả dữ liệu cũ bằng dữ liệu mới
		if err := s.volumeRepo.ReplaceAllForSymbol(symbol, records); err != nil {
			fmt.Printf("Lỗi lưu DB %s: %v\n", symbol, err)
		} else {
			fmt.Printf("Đã cập nhật %d records volume cho %s\n", len(records), symbol)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (s *AutoVolumeService) AnalyzeAndNotifyVolumes(channelID string) error {
	// Lấy tất cả symbols thay vì tất cả records
	symbols, err := s.symbolRepo.GetAllSymbols()
	if err != nil {
		return err
	}
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
		records22, _ := s.volumeRepo.GetLastNBySymbol(symbol, 23)
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
		for _, r := range records22[1:] {
			totalCandlestickLength += r.CandlestickLength()
			totalCandlestickBody += r.CandlestickBody()
		}
		averageCandlestickBody := totalCandlestickBody / float64(len(records22)-1)

		volumeAnalysis := taService.analyzeVolumeFromFloat64(volumes)
		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			// Lấy bản ghi MỚI NHẤT (records22[0])
			latestRecord := records22[0]
			// lấy bản ghi cây nến thứ 21
			record21 := records22[1]
			// lấy bản ghi cây nến thứ 20
			record20 := records22[2]

			// Lấy time hiện tại
			currentTime := time.Now().In(loc)
			formattedTime := currentTime.Format("2006-01-02 15:04:05")

			// Phân tích mô hình
			engulfingResult := detectEngulfing(record20, record21)
			confirmation1 := engulfingResult.Confirmation
			pattern1 := engulfingResult.Pattern
			piercingResult := detectPiercingPattern(record20, record21, averageCandlestickBody)
			confirmation2 := piercingResult.Confirmation
			pattern2 := piercingResult.Pattern

			message := fmt.Sprintf("💰*[ALERT]* Symbol: %s\n"+
				"📅 Time: %s\n"+
				"🚀Volume: %s\n"+
				"🚀SMA21: %s\n"+
				"🚀Price: %s\n"+
				"🎯Strength: %s\n"+
				"🔥Signal: %s\n"+
				"🔥Pattern: %s %s\n"+
				"🔥Confirmation: %s\n %s\n"+
				"Check record21 %d , record20 %d",
				strings.TrimSuffix(latestRecord.Symbol, "USDT"),
				formattedTime,
				utils.FormatVolume(decimal.NewFromFloat(latestRecord.QuoteAssetVolume)),
				utils.FormatVolume(volumeAnalysis.VolumeSMA21),
				utils.FormatPrice(decimal.NewFromFloat(latestRecord.ClosePrice)),
				volumeAnalysis.VolumeStrength,
				volumeAnalysis.VolumeSignal,
				pattern1, pattern2,
				confirmation1, confirmation2,
				record21.ID, record20.ID,
			)
			s.telegramBotService.SendTelegramToChannel(channelID, message)
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
}

func detectEngulfing(record20, record21 models.AutoVolumeRecord) PatternDetectionResult {
	if record20.Candlestick() == 0 &&
		record21.Candlestick() == 1 &&
		record21.QuoteAssetVolume > record20.QuoteAssetVolume*1.2 &&
		record21.OpenPrice < record20.ClosePrice &&
		record21.ClosePrice > record20.OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình Bullish Engulfing",
			Confirmation: "✅ Đây là một tín hiệu đảo chiều tăng giá rất mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng giảm. Nó cho thấy phe mua đã hoàn toàn áp đảo phe bán",
			IsDetected:   true,
		}
	} else if record20.Candlestick() == 1 &&
		record21.Candlestick() == 0 &&
		record21.QuoteAssetVolume > record20.QuoteAssetVolume*1.2 &&
		record21.OpenPrice > record20.ClosePrice &&
		record21.ClosePrice < record20.OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình Bearish Engulfing",
			Confirmation: "❎ Đây là một tín hiệu đảo chiều giảm giá mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng tăng. Nó cho thấy phe bán đã hoàn toàn áp đảo phe mua",
			IsDetected:   true,
		}
	}
	return PatternDetectionResult{IsDetected: false}
}

func detectPiercingPattern(record20, record21 models.AutoVolumeRecord, averageCandlestickBody float64) PatternDetectionResult {
	if record20.Candlestick() == 0 &&
		record20.IsCandlestickBodyLong(averageCandlestickBody, 1.5) &&
		record21.Candlestick() == 1 &&
		record21.OpenPrice < record20.ClosePrice && // Nến 2 mở cửa dưới giá đóng cửa nến 1 (có thể mở dưới cả low)
		record21.ClosePrice > record20.CandlestBodyMidpoint() && // Nến 2 đóng cửa trên điểm giữa thân nến 1
		record21.ClosePrice < record20.OpenPrice { // Nến 2 đóng cửa dưới giá mở cửa nến 1 (không phải nhấn chìm) {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình Piercing Pattern",
			Confirmation: "✅ Tín hiệu đảo chiều tăng giá. Phe mua đã giành lại quyền kiểm soát sau một đợt giảm giá mạnh",
			IsDetected:   true,
		}
	}
	return PatternDetectionResult{IsDetected: false}
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
		next := now.Truncate(time.Hour).Add(time.Hour)
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
		// Cắt lẻ đến giờ, sau đó thêm 1 giờ + 2 phút (ví dụ: 8:30 → 9:02:00)
		next := now.Truncate(time.Hour).Add(time.Hour + 2*time.Minute)
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
