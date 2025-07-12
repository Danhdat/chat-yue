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
		for _, k := range recentKlines {
			quoteAssetVolumeStr := k[7].(string)
			quoteAssetVolume, _ := strconv.ParseFloat(quoteAssetVolumeStr, 64)

			record := &models.AutoVolumeRecord{
				Symbol:           symbol,
				QuoteAssetVolume: quoteAssetVolume,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			if err := s.volumeRepo.Upsert(record); err != nil {
				fmt.Printf("Lỗi lưu DB %s: %v\n", symbol, err)
			} else {
				fmt.Printf("Đã lưu volume cho %s: %f\n", symbol, quoteAssetVolume)
			}
		}
		time.Sleep(1 * time.Second)
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

	for _, symbol := range symbols {
		// Kiểm tra nếu symbol đã được xử lý
		if processedSymbols[symbol] {
			continue
		}

		records22, _ := s.volumeRepo.GetLastNBySymbol(symbol, 22)
		log.Println("records22 for symbol ", symbol, ": ", len(records22), "records")

		// Kiểm tra nếu không có dữ liệu
		if len(records22) == 0 {
			continue
		}

		var volumes []float64
		for _, r := range records22 {
			volumes = append(volumes, r.QuoteAssetVolume)
		}

		volumeAnalysis := taService.analyzeVolumeFromFloat64(volumes)
		log.Println("Volume analysis for ", symbol, ": ", volumeAnalysis.VolumeStrength)

		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			// Lấy bản ghi MỚI NHẤT (records22[0])
			latestRecord := records22[0]
			log.Println("latestRecord: ", latestRecord)

			message := fmt.Sprintf("[ALERT] Symbol: %s\nVolume: %s\nStrength: %s\nSignal: %s", latestRecord.Symbol, utils.FormatVolume(decimal.NewFromFloat(latestRecord.QuoteAssetVolume)), volumeAnalysis.VolumeStrength, volumeAnalysis.VolumeSignal)
			s.telegramBotService.SendTelegramToChannel(channelID, message)
			log.Printf("Đã gửi cảnh báo volume cho %s", latestRecord.Symbol)
		}

		// Đánh dấu symbol đã được xử lý
		processedSymbols[symbol] = true
		time.Sleep(1 * time.Second)
	}

	return nil
}

// Hàm phân tích volume cho 1 giá trị float64 (tương thích với analyzeVolume)
func (s *TechnicalAnalysisService) analyzeVolumeFromFloat64(volumes []float64) models.VolumeAnalysis {
	if len(volumes) < models.VOLUME_SMA_PERIOD+1 {
		return models.VolumeAnalysis{}
	}
	// Chuyển sang decimal.Decimal để dùng lại logic cũ nếu cần
	currentVolume := decimal.NewFromFloat(volumes[len(volumes)-1])
	var sum float64
	for i := len(volumes) - models.VOLUME_SMA_PERIOD - 1; i < len(volumes)-1; i++ {
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
	if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(3.0)) {
		volumeSignal = "🔥 VOLUME EXPLOSION"
		volumeStrength = "EXTREME"
		confirmation = "Tín hiệu Cực MẠNH - Breakout/Breakdown được xác nhận"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(2.0)) {
		volumeSignal = "🚀 HIGH VOLUME SPIKE"
		volumeStrength = "STRONG"
		confirmation = "Tín hiệu MẠNH - Xu hướng được hỗ trợ tốt"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(1.5)) {
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
	go s.Run()
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
		case <-time.After(4*time.Hour + 10*time.Minute):
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
