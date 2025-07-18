package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Constants cho các chỉ báo kỹ thuật
const (
	RSI_PERIOD = 14
	EMA_PERIOD = 20

	// 3 EMA System
	EMA_SHORT  = 9  // EMA ngắn hạn
	EMA_MEDIUM = 21 // EMA trung hạn
	EMA_LONG   = 50 // EMA dài hạn

	// Volume Analysis
	VOLUME_SMA_PERIOD = 21  // SMA của Volume (21 kỳ)
	VOLUME_SPIKE_1_5X = 1.5 // Volume spike 1.5x
	VOLUME_SPIKE_2X   = 2.0 // Volume spike 2x
	VOLUME_SPIKE_3X   = 3.0 // Volume spike 3x
)

// CryptoPrice đại diện cho giá của một cryptocurrency
type CryptoPrice struct {
	ID                       int64           `json:"id"`
	Symbol                   string          `json:"symbol"`
	Name                     string          `json:"name"`
	CurrentPrice             decimal.Decimal `json:"current_price"`
	MarketCap                decimal.Decimal `json:"market_cap"`
	Volume24h                decimal.Decimal `json:"volume_24h"`
	PriceChange24h           decimal.Decimal `json:"price_change_24h"`
	PriceChangePercentage24h decimal.Decimal `json:"price_change_percentage_24h"`
	LastUpdated              time.Time       `json:"last_updated"`
}

// KlineData đại diện cho một nến từ Binance API
type KlineData struct {
	OpenTime                 int64  `json:"open_time"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"close_time"`
	QuoteAssetVolume         string `json:"quote_asset_volume"`
	NumberOfTrades           int    `json:"number_of_trades"`
	TakerBuyBaseAssetVolume  string `json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume string `json:"taker_buy_quote_asset_volume"`
	Ignore                   string `json:"ignore"`
}

// PriceData đại diện cho dữ liệu giá theo thời gian
type PriceData struct {
	Timestamp time.Time       `json:"timestamp"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
}

// TechnicalIndicators chứa các chỉ báo kỹ thuật
type TechnicalIndicators struct {
	ID            int64           `json:"id"`
	Symbol        string          `json:"symbol"`
	RSI           decimal.Decimal `json:"rsi"`
	EMA20         decimal.Decimal `json:"ema_20"`
	EMA50         decimal.Decimal `json:"ema_50"`
	EMA200        decimal.Decimal `json:"ema_200"`
	MACD          decimal.Decimal `json:"macd"`
	MACDSignal    decimal.Decimal `json:"macd_signal"`
	MACDHistogram decimal.Decimal `json:"macd_histogram"`
	Timestamp     time.Time       `json:"timestamp"`
}

// UserSession lưu trữ phiên người dùng
type UserSession struct {
	UserID             int64           `json:"user_id"`
	ChatID             int64           `json:"chat_id"`
	Username           string          `json:"username"`
	PreferredSymbols   []string        `json:"preferred_symbols"`
	AlertRSIOverbought decimal.Decimal `json:"alert_rsi_overbought"`
	AlertRSIOversold   decimal.Decimal `json:"alert_rsi_oversold"`
	LastActive         time.Time       `json:"last_active"`
	CreatedAt          time.Time       `json:"created_at"`
}

type TrendAnalysis struct {
	Direction      string // "bullish", "bearish", "sideways"
	Strength       string // "strong", "moderate", "weak"
	Signals        []string
	Recommendation string
}

// Volume Analysis Structure
type VolumeAnalysis struct {
	CurrentVolume  decimal.Decimal
	VolumeSMA21    decimal.Decimal
	VolumeRatio    decimal.Decimal
	VolumeSignal   string
	VolumeStrength string
	Confirmation   string
}

// AnalysisData chứa dữ liệu phân tích chi tiết
type AnalysisData struct {
	Symbol         string
	Interval       string
	CurrentPrice   float64
	RSI            float64
	EMA9           float64
	EMA21          float64
	EMA50          float64
	MACD           float64
	MACDSignal     float64
	VolumeSMA      float64
	Trend          string
	Power          string
	Signal         string
	Recommendation string
	VolumeSignal   string
	VolumeAnalysis VolumeAnalysis
}
