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

type SymbolRepository struct {
	db *gorm.DB
}

func NewSymbolRepository() *SymbolRepository {
	return &SymbolRepository{db: DB}
}

func (r *SymbolRepository) Create(symbol *Symbol) error {
	return r.db.Create(symbol).Error
}

func (r *SymbolRepository) UpdateLastUpdateTime() error {
	var dataUpdate DataUpdate
	// tìm hoặc tạo record
	result := r.db.Model(&DataUpdate{}).Where("table_name = ?", "symbols").First(&dataUpdate)
	if result.Error != nil {
		dataUpdate := DataUpdate{
			TableName:  "symbols",
			LastUpdate: time.Now(),
		}
		return r.db.Create(&dataUpdate).Error
	}
	// cập nhật thời gian
	dataUpdate.LastUpdate = time.Now()
	return r.db.Save(&dataUpdate).Error
}

func (r *SymbolRepository) GetAllSymbols() ([]string, error) {
	var symbols []Symbol
	err := r.db.Find(&symbols).Error
	if err != nil {
		return nil, err
	}
	var result []string
	for _, s := range symbols {
		result = append(result, s.Symbol)
	}
	return result, nil
}

func (r *SymbolRepository) GetSymbolByBaseAsset(baseAsset string) ([]Symbol, error) {
	var symbols []Symbol
	err := r.db.Where("base_asset = ?", baseAsset).First(&symbols).Error
	return symbols, err
}

func (r *SymbolRepository) SaveToDatabase(symbols []Symbol) error {
	// Xoá dữ liệu cũ
	if err := r.db.Unscoped().Where("1 = 1").Delete(&Symbol{}).Error; err != nil {
		return err
	}
	// Lưu dữ liệu mới
	if len(symbols) > 0 {
		if err := r.db.Create(&symbols).Error; err != nil {
			return err
		}
	}
	return nil
}

const updateInterval = 15 * 24 * time.Hour // 15 ngày
func (r *SymbolRepository) ShouldUpdate() bool {
	var dataUpdate DataUpdate
	err := r.db.Model(&DataUpdate{}).Where("table_name = ?", "symbols").First(&dataUpdate).Error
	if err != nil {
		return true
	}
	return time.Since(dataUpdate.LastUpdate) > updateInterval
}

type AutoVolumeRecordRepository struct {
	db *gorm.DB
}

func NewAutoVolumeRecordRepository() *AutoVolumeRecordRepository {
	return &AutoVolumeRecordRepository{db: DB}
}

func (r *AutoVolumeRecordRepository) Create(record *AutoVolumeRecord) error {
	return r.db.Create(record).Error
}

// ReplaceAllForSymbol xóa tất cả dữ liệu cũ của symbol và thêm dữ liệu mới
func (r *AutoVolumeRecordRepository) ReplaceAllForSymbol(symbol string, records []AutoVolumeRecord) error {
	// Bắt đầu transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Xóa tất cả dữ liệu cũ của symbol
	if err := tx.Unscoped().Where("symbol = ?", symbol).Delete(&AutoVolumeRecord{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Thêm dữ liệu mới
	if len(records) > 0 {
		if err := tx.Create(&records).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

func (r *AutoVolumeRecordRepository) GetLastNBySymbol(symbol string, n int) ([]AutoVolumeRecord, error) {
	var records []AutoVolumeRecord
	err := r.db.Where("symbol = ?", symbol).Order("created_at DESC").Limit(n).Find(&records).Error
	return records, err
}
