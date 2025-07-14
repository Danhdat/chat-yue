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
		url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=4h&limit=22", symbol)
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
		// Lấy 22 nến gần nhất
		recentKlines := klines
		if len(klines) > 22 {
			recentKlines = klines[len(klines)-22:]
		}

		loc := time.FixedZone("UTC+7", 7*60*60)

		// Tạo slice để lưu tất cả records cho symbol này
		var records []models.AutoVolumeRecord

		for _, k := range recentKlines {
			quoteAssetVolumeStr := k[7].(string)
			quoteAssetVolume, _ := strconv.ParseFloat(quoteAssetVolumeStr, 64)

			record := models.AutoVolumeRecord{
				Symbol:           symbol,
				QuoteAssetVolume: quoteAssetVolume,
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
	symbolSendCount := make(map[string]int)
	loc := time.FixedZone("UTC+7", 7*60*60)

	for _, symbol := range symbols {
		// Kiểm tra nếu symbol đã được xử lý
		if processedSymbols[symbol] {
			continue
		}
		records22, _ := s.volumeRepo.GetLastNBySymbol(symbol, 22)
		// Kiểm tra nếu không có dữ liệu
		if len(records22) == 0 {
			continue
		}

		var volumes []float64
		for _, r := range records22 {
			volumes = append(volumes, r.QuoteAssetVolume)
		}

		volumeAnalysis := taService.analyzeVolumeFromFloat64(volumes)
		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			// Lấy bản ghi MỚI NHẤT (records22[0])
			latestRecord := records22[0]

			// Lấy time hiện tại
			currentTime := time.Now().In(loc)
			formattedTime := currentTime.Format("2006-01-02 15:04:05")

			//Đếm số lần gửi của symbol
			symbolSendCount[latestRecord.Symbol]++
			count := symbolSendCount[latestRecord.Symbol]

			message := fmt.Sprintf("💰[ALERT] Symbol: %s\n"+
				"📅 Time: %s\n"+
				"🔄 Occurrences: %d\n"+
				"🚀Volume: %s\n"+
				"🚀SMA21: %s\n"+
				"🎯Strength: %s\n"+
				"🔥Signal: %s\n"+
				"🔥Confirmation: %s",
				latestRecord.Symbol,
				formattedTime,
				count,
				utils.FormatVolume(decimal.NewFromFloat(latestRecord.QuoteAssetVolume)),
				utils.FormatVolume(volumeAnalysis.VolumeSMA21),
				volumeAnalysis.VolumeStrength,
				volumeAnalysis.VolumeSignal,
				volumeAnalysis.Confirmation)
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
	var volumeSignal, volumeStrength, confirmation string
	var volumeRatio decimal.Decimal
	if volumeSMA > 0 {
		volumeRatio = currentVolume.Div(decimal.NewFromFloat(volumeSMA))
	} else {
		volumeRatio = decimal.Zero
	}
	if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_3X)) {
		volumeSignal = "🔥 VOLUME EXPLOSION"
		volumeStrength = "EXTREME"
		confirmation = "Tín hiệu Cực MẠNH - Breakout/Breakdown được xác nhận"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_2X)) {
		volumeSignal = "🚀 HIGH VOLUME SPIKE"
		volumeStrength = "STRONG"
		confirmation = "Tín hiệu MẠNH - Xu hướng được hỗ trợ tốt"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_1_5X)) {
		volumeSignal = "📈 ABOVE AVERAGE VOLUME"
		volumeStrength = "MODERATE"
		confirmation = "Tín hiệu TRUNG BÌNH - Có sự quan tâm tăng lên"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(1.0)) {
		volumeSignal = "🟡 NORMAL VOLUME"
		volumeStrength = "NORMAL"
		confirmation = "Volume bình thường - Theo dõi thêm"
	} else {
		volumeSignal = "📉 LOW VOLUME"
		volumeStrength = "WEAK"
		confirmation = "Volume thấp - Tín hiệu yếu, cẩn thận với fake move"
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
	// Chạy cập nhật định kỳ mỗi 4 giờ
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go s.Run()
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
	go s.Run()
	for {
		select {
		case <-time.After(4*time.Hour + 5*time.Minute):
			go s.Run()
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
