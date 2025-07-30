package models

import (
	"fmt"
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

type CommonRepository struct {
	db *gorm.DB
}

func NewCommonRepository() *CommonRepository {
	return &CommonRepository{db: DB}
}
func (r *CommonRepository) UpdateLastUpdateTime(tableName string) error {
	var dataUpdate DataUpdate
	result := r.db.Model(&DataUpdate{}).Where("table_name = ?", tableName).First(&dataUpdate)
	if result.Error != nil {
		return r.db.Create(&DataUpdate{
			TableName:  tableName,
			LastUpdate: time.Now(),
		}).Error
	}
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
func (r *CommonRepository) ShouldUpdate(tableName string) bool {
	var dataUpdate DataUpdate
	err := r.db.Model(&DataUpdate{}).Where("table_name = ?", tableName).First(&dataUpdate).Error
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
	err := r.db.Where("symbol = ?", symbol).Order("open_time DESC").Limit(n).Find(&records).Error
	return records, err
}

// NotificationLogRepository xử lý thao tác với bảng notification_logs
type NotificationLogRepository struct {
	db *gorm.DB
}

// NewNotificationLogRepository tạo instance mới
func NewNotificationLogRepository() *NotificationLogRepository {
	return &NotificationLogRepository{db: DB}
}

// Create lưu log thông báo mới - Tối ưu hóa
func (r *NotificationLogRepository) Create(log *NotificationLog) error {
	// Sử dụng Select để chỉ insert các field cần thiết
	return r.db.Select("symbol", "created_at", "direction", "type").Create(log).Error
}

// CreateBatch lưu nhiều log thông báo cùng lúc - Tối ưu hóa cho batch insert
func (r *NotificationLogRepository) CreateBatch(logs []*NotificationLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Sử dụng batch insert với size 100
	batchSize := 100
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}

		batch := logs[i:end]
		if err := r.db.Select("symbol", "created_at", "direction", "type").CreateInBatches(batch, len(batch)).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateAsync lưu log thông báo bất đồng bộ - Tối ưu hóa cho performance
func (r *NotificationLogRepository) CreateAsync(log *NotificationLog) {
	go func() {
		if err := r.Create(log); err != nil {
			// Log error nhưng không block main thread
			// Sử dụng fmt.Printf thay vì log.Printf để tránh conflict với parameter name
			fmt.Printf("Lỗi lưu log thông báo bất đồng bộ: %v\n", err)
		}
	}()
}

// CountBySymbolToday đếm số lần gửi tin nhắn cho một symbol trong ngày hôm nay
func (r *NotificationLogRepository) CountBySymbolToday(symbol string) (int64, error) {
	var count int64
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	err := r.db.Model(&NotificationLog{}).
		Where("symbol = ? AND created_at >= ?", symbol, today).
		Count(&count).Error
	return count, err
}

// đếm số lần gửi tin nhắn cho một symbol trong một tuần
func (r *NotificationLogRepository) CountBySymbolThisWeek(symbol string) (int64, error) {
	var count int64
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	year, week := now.ISOWeek()

	// Tính thứ 2 đầu tuần và chủ nhật cuối tuần
	startOfWeek := firstDayOfISOWeek(year, week, loc)
	endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.Model(&NotificationLog{}).
		Where("symbol = ? AND created_at >= ? AND created_at <= ?",
			symbol,
			startOfWeek,
			endOfWeek,
		).
		Count(&count).Error

	return count, err
}

// Lấy toàn bộ NotificationLog trong tuần hiện tại
func (r *NotificationLogRepository) GetLogsThisWeek() ([]NotificationLog, error) {
	loc := time.FixedZone("UTC+7", 7*60*60)
	now := time.Now().In(loc)
	year, week := now.ISOWeek()
	startOfWeek := firstDayOfISOWeek(year, week, loc)
	endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var logs []NotificationLog
	err := r.db.Where("created_at >= ? AND created_at <= ?", startOfWeek, endOfWeek).Find(&logs).Error
	return logs, err
}

func firstDayOfISOWeek(year, week int, loc *time.Location) time.Time {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, 1)
	}
	isoYear, isoWeek := date.ISOWeek()
	for isoYear < year || isoWeek < week {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}
func (r *NotificationLogRepository) DeleteLogMonth() error {
	return r.db.Where("created_at < ?", time.Now().AddDate(0, -1, 0)).Delete(&NotificationLog{}).Error
}

// ALPHA BINANCE
type AlphaSymbolRepository struct {
	db *gorm.DB
}

func NewAlphaSymbolRepository() *AlphaSymbolRepository {
	return &AlphaSymbolRepository{db: DB}
}

func (r *AlphaSymbolRepository) SaveToDatabaseAlpha(symbols []AlphaSymbol) error {
	// Xoá dữ liệu cũ
	if err := r.db.Unscoped().Where("1 = 1").Delete(&AlphaSymbol{}).Error; err != nil {
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

func (r *AlphaSymbolRepository) GetAllAlphaSymbols() ([]string, error) {
	var symbols []AlphaSymbol
	err := r.db.Find(&symbols).Error
	if err != nil {
		return nil, err
	}
	var result []string
	for _, s := range symbols {
		if !s.CexOffDisplay {
			result = append(result, s.AlphaID)
		}
	}
	return result, nil
}

func (r *AlphaSymbolRepository) GetAllAlphaName() ([]string, error) {
	var symbols []AlphaSymbol
	err := r.db.Find(&symbols).Error
	if err != nil {
		return nil, err
	}
	var result []string
	for _, s := range symbols {
		if !s.CexOffDisplay {
			result = append(result, s.Symbol)
		}
	}
	return result, nil
}

func (r *AlphaSymbolRepository) GetNameByAlphaSymbol(symbol string) (string, error) {
	var alphaSymbol AlphaSymbol
	err := r.db.Where("alpha_id = ?", symbol).First(&alphaSymbol).Error
	return alphaSymbol.Symbol, err
}
