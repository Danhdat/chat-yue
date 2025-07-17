package services

import (
	"chatbtc/models"
	"log"
	"time"
)

// AnalysisService xử lý các thao tác liên quan đến phân tích
type AnalysisService struct {
	analysisRepo *models.AnalysisRepository
	priceRepo    *models.PriceHistoryRepository
}

// NewAnalysisService tạo instance mới
func NewAnalysisService() *AnalysisService {
	return &AnalysisService{
		analysisRepo: models.NewAnalysisRepository(),
		priceRepo:    models.NewPriceHistoryRepository(),
	}
}

// SaveAnalysis lưu kết quả phân tích vào database
func (s *AnalysisService) SaveAnalysis(symbol, interval string, closePrice, volume, rsi, ema9, ema21, ema50, macd, macdSignal, volumeSMA float64, trend, power, signal, recommendation, volumeSignal string) error {
	loc := time.FixedZone("UTC+7", 7*60*60)
	record := &models.AnalysisRecord{
		Symbol:         symbol,
		Interval:       interval,
		ClosePrice:     closePrice,
		Volume:         volume,
		RSI:            rsi,
		EMA9:           ema9,
		EMA21:          ema21,
		EMA50:          ema50,
		MACD:           macd,
		MACDSignal:     macdSignal,
		VolumeSMA:      volumeSMA,
		Trend:          trend,
		Power:          power,
		Signal:         signal,
		Recommendation: recommendation,
		VolumeSignal:   volumeSignal,
		CreatedAt:      time.Now().In(loc),
		UpdatedAt:      time.Now().In(loc),
	}

	err := s.analysisRepo.Create(record)
	if err != nil {
		log.Printf("❌ Lỗi lưu phân tích: %v", err)
		return err
	}

	log.Printf("✅ Đã lưu phân tích cho %s (%s)", symbol, interval)
	return nil
}

// GetAnalysisHistory lấy lịch sử phân tích
func (s *AnalysisService) GetAnalysisHistory(symbol, interval string, limit int) ([]models.AnalysisRecord, error) {
	return s.analysisRepo.GetBySymbolAndInterval(symbol, interval, limit)
}
