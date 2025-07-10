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
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.analysisRepo.Create(record)
	if err != nil {
		log.Printf("❌ Lỗi lưu phân tích: %v", err)
		return err
	}

	log.Printf("✅ Đã lưu phân tích cho %s (%s)", symbol, interval)
	return nil
}

// SavePriceHistory lưu lịch sử giá vào database
func (s *AnalysisService) SavePriceHistory(symbol, interval string, openTime time.Time, open, high, low, close, volume float64) error {
	// Kiểm tra xem nến này đã tồn tại chưa
	existingHistory, err := s.priceRepo.GetBySymbolAndInterval(symbol, interval, 1)
	if err == nil && len(existingHistory) > 0 {
		latestHistory := existingHistory[0]
		// Nếu nến mới có cùng thời gian với nến cuối cùng, không lưu
		if latestHistory.OpenTime.Equal(openTime) {
			log.Printf("ℹ️ Nến %s (%s) tại %s đã tồn tại, bỏ qua", symbol, interval, openTime.Format("2006-01-02 15:04:05"))
			return nil
		}
	}

	history := &models.PriceHistory{
		Symbol:    symbol,
		Interval:  interval,
		OpenTime:  openTime,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.priceRepo.Create(history)
	if err != nil {
		log.Printf("❌ Lỗi lưu lịch sử giá: %v", err)
		return err
	}

	log.Printf("✅ Đã lưu nến mới: %s (%s) tại %s", symbol, interval, openTime.Format("2006-01-02 15:04:05"))
	return nil
}

// GetLatestAnalysis lấy phân tích mới nhất
func (s *AnalysisService) GetLatestAnalysis(symbol, interval string) (*models.AnalysisRecord, error) {
	return s.analysisRepo.GetLatestAnalysis(symbol, interval)
}

// GetAnalysisHistory lấy lịch sử phân tích
func (s *AnalysisService) GetAnalysisHistory(symbol, interval string, limit int) ([]models.AnalysisRecord, error) {
	return s.analysisRepo.GetBySymbolAndInterval(symbol, interval, limit)
}

// GetPriceHistory lấy lịch sử giá
func (s *AnalysisService) GetPriceHistory(symbol, interval string, limit int) ([]models.PriceHistory, error) {
	return s.priceRepo.GetBySymbolAndInterval(symbol, interval, limit)
}

// GetPriceHistoryCount lấy số lượng record price history
func (s *AnalysisService) GetPriceHistoryCount(symbol, interval string) (int64, error) {
	return s.priceRepo.GetCount(symbol, interval)
}

// CleanOldPriceData dọn dẹp dữ liệu giá cũ
func (s *AnalysisService) CleanOldPriceData(symbol, interval string, keepDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	err := s.priceRepo.DeleteOldData(symbol, interval, cutoffTime)
	if err != nil {
		log.Printf("❌ Lỗi dọn dẹp dữ liệu cũ: %v", err)
		return err
	}
	log.Printf("✅ Đã dọn dẹp dữ liệu %s (%s) cũ hơn %d ngày", symbol, interval, keepDays)
	return nil
}
