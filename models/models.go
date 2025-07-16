package models

import (
	"math"
	"time"

	"gorm.io/gorm"
)

// AnalysisRecord lưu trữ lịch sử phân tích
type AnalysisRecord struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Symbol         string         `gorm:"not null;index" json:"symbol"`
	Interval       string         `gorm:"not null" json:"interval"`
	ClosePrice     float64        `gorm:"not null" json:"close_price"`
	Volume         float64        `gorm:"not null" json:"volume"`
	RSI            float64        `json:"rsi"`
	EMA9           float64        `json:"ema_9"`
	EMA21          float64        `json:"ema_21"`
	EMA50          float64        `json:"ema_50"`
	MACD           float64        `json:"macd"`
	MACDSignal     float64        `json:"macd_signal"`
	VolumeSMA      float64        `json:"volume_sma"`
	Trend          string         `json:"trend"`
	Power          string         `json:"power"`
	Signal         string         `json:"signal"`
	Recommendation string         `json:"recommendation"`
	VolumeSignal   string         `json:"volume_signal"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// PriceHistory lưu trữ lịch sử giá
type PriceHistory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Symbol    string         `gorm:"not null;index" json:"symbol"`
	Interval  string         `gorm:"not null" json:"interval"`
	OpenTime  time.Time      `gorm:"not null" json:"open_time"`
	Open      float64        `gorm:"not null" json:"open"`
	High      float64        `gorm:"not null" json:"high"`
	Low       float64        `gorm:"not null" json:"low"`
	Close     float64        `gorm:"not null" json:"close"`
	Volume    float64        `gorm:"not null" json:"volume"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName định nghĩa tên bảng cho AnalysisRecord
func (AnalysisRecord) TableName() string {
	return "analysis_records"
}

// TableName định nghĩa tên bảng cho PriceHistory
func (PriceHistory) TableName() string {
	return "price_histories"
}

type AutoVolumeRecord struct {
	ID               uint    `gorm:"primaryKey"`
	Symbol           string  `gorm:"index;not null"`
	QuoteAssetVolume float64 `gorm:"not null"`
	OpenPrice        float64 `gorm:"not null"`
	ClosePrice       float64 `gorm:"not null"`
	HighPrice        float64 `gorm:"not null"`
	LowPrice         float64 `gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (AutoVolumeRecord) TableName() string {
	return "auto_volume_record"
}

func (r *AutoVolumeRecord) Candlestick() float64 {
	// 1: green, 0: red
	if r.ClosePrice > r.OpenPrice {
		return 1
	} else {
		return 0
	}
}

func (r *AutoVolumeRecord) CandlestickBody() float64 {
	return math.Abs(r.ClosePrice - r.OpenPrice)
}

func (r *AutoVolumeRecord) CandlestickLength() float64 {
	return r.HighPrice - r.LowPrice
}

func (r *AutoVolumeRecord) CandlestickUpperShadow() float64 {
	if r.ClosePrice > r.OpenPrice {
		return r.HighPrice - r.ClosePrice
	} else {
		return r.HighPrice - r.OpenPrice
	}
}

func (r *AutoVolumeRecord) CandlestickLowerShadow() float64 {
	if r.ClosePrice > r.OpenPrice {
		return r.OpenPrice - r.LowPrice
	} else {
		return r.ClosePrice - r.LowPrice
	}
}

func (r *AutoVolumeRecord) IsCandlestickBodyLong(avgLength float64, multiplier float64) bool {
	return r.CandlestickBody() > avgLength*multiplier
}

func (r *AutoVolumeRecord) IsCandlestickBodyShort(avgLength float64, multiplier float64) bool {
	return r.CandlestickBody() < avgLength*multiplier
}

// Tính toán điểm ở giữa cây nến
func (r *AutoVolumeRecord) CandlestBodyMidpoint() float64 {
	if r.OpenPrice >= r.ClosePrice { // green candlestick
		return r.OpenPrice + (r.ClosePrice-r.OpenPrice)/2
	} else { // red candlestick
		return r.ClosePrice + (r.OpenPrice-r.ClosePrice)/2
	}
}
