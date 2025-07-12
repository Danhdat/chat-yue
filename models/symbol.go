package models

import (
	"time"

	"gorm.io/gorm"
)

// Symbol represents a trading pair from Binance
type Symbol struct {
	ID        uint   `gorm:"primaryKey"`
	Symbol    string `gorm:"unique;not null"`
	Status    string `gorm:"not null"`
	BaseAsset string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DataUpdate tracks when data was last updated
type DataUpdate struct {
	ID         uint      `gorm:"primaryKey"`
	TableName  string    `gorm:"unique;not null"`
	LastUpdate time.Time `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// BinanceExchangeInfo represents the structure from Binance API
type BinanceExchangeInfo struct {
	Symbols []BinanceSymbol `json:"symbols"`
}

// BinanceSymbol represents a symbol from Binance API
type BinanceSymbol struct {
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}

// BeforeCreate will set timestamps
func (s *Symbol) BeforeCreate(tx *gorm.DB) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate will update the updated_at field
func (s *Symbol) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
