package models

import (
	"time"

	"gorm.io/gorm"
)

// AnalysisRepository xử lý thao tác với bảng analysis_records
type AnalysisRepository struct {
	db *gorm.DB
}

// NewAnalysisRepository tạo instance mới
func NewAnalysisRepository() *AnalysisRepository {
	return &AnalysisRepository{db: DB}
}

// Create lưu record phân tích mới
func (r *AnalysisRepository) Create(record *AnalysisRecord) error {
	return r.db.Create(record).Error
}

// GetBySymbolAndInterval lấy lịch sử phân tích theo symbol và interval
func (r *AnalysisRepository) GetBySymbolAndInterval(symbol, interval string, limit int) ([]AnalysisRecord, error) {
	var records []AnalysisRecord
	err := r.db.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// GetLatestAnalysis lấy phân tích mới nhất
func (r *AnalysisRepository) GetLatestAnalysis(symbol, interval string) (*AnalysisRecord, error) {
	var record AnalysisRecord
	err := r.db.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// PriceHistoryRepository xử lý thao tác với bảng price_histories
type PriceHistoryRepository struct {
	db *gorm.DB
}

// NewPriceHistoryRepository tạo instance mới
func NewPriceHistoryRepository() *PriceHistoryRepository {
	return &PriceHistoryRepository{db: DB}
}

// Create lưu lịch sử giá mới
func (r *PriceHistoryRepository) Create(history *PriceHistory) error {
	return r.db.Create(history).Error
}

// GetBySymbolAndInterval lấy lịch sử giá theo symbol và interval
func (r *PriceHistoryRepository) GetBySymbolAndInterval(symbol, interval string, limit int) ([]PriceHistory, error) {
	var histories []PriceHistory
	err := r.db.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("open_time DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// GetLatestPrice lấy giá mới nhất
func (r *PriceHistoryRepository) GetLatestPrice(symbol, interval string) (*PriceHistory, error) {
	var history PriceHistory
	err := r.db.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("open_time DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// DeleteOldData xóa dữ liệu cũ hơn một khoảng thời gian
func (r *PriceHistoryRepository) DeleteOldData(symbol, interval string, olderThan time.Time) error {
	return r.db.Where("symbol = ? AND interval = ? AND open_time < ?", symbol, interval, olderThan).
		Delete(&PriceHistory{}).Error
}

// GetCount lấy số lượng record cho symbol và interval
func (r *PriceHistoryRepository) GetCount(symbol, interval string) (int64, error) {
	var count int64
	err := r.db.Model(&PriceHistory{}).
		Where("symbol = ? AND interval = ?", symbol, interval).
		Count(&count).Error
	return count, err
}
