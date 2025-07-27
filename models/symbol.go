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

// AlphaAPIResponse represents the top-level structure of the Alpha API response.
type AlphaAPIResponse struct {
	Code    string        `json:"code"`
	Message interface{}   `json:"message"` // Can be null, so interface{} is safer
	Data    []AlphaSymbol `json:"data"`
}

// AlphaSymbol represents a token from the Alpha API.
// Note: Numerical values are stored as strings as the API returns them that way.
type AlphaSymbol struct {
	ID                uint   `gorm:"primaryKey"`
	TokenID           string `gorm:"unique;not null" json:"tokenId"`
	ChainID           string `gorm:"not null" json:"chainId"`
	ChainName         string `json:"chainName"`
	ContractAddress   string `gorm:"unique;not null" json:"contractAddress"`
	Symbol            string `gorm:"not null" json:"symbol"`
	PercentChange24h  string `gorm:"type:varchar(255)" json:"percentChange24h"`
	Volume24h         string `gorm:"type:varchar(255)" json:"volume24h"`
	MarketCap         string `gorm:"type:varchar(255)" json:"marketCap"`
	FDV               string `gorm:"type:varchar(255)" json:"fdv"`
	Liquidity         string `gorm:"type:varchar(255)" json:"liquidity"`
	TotalSupply       string `gorm:"type:varchar(255)" json:"totalSupply"`
	CirculatingSupply string `gorm:"type:varchar(255)" json:"circulatingSupply"`
	Holders           string `json:"holders"`
	ListingCex        bool   `json:"listingCex"`
	HotTag            bool   `json:"hotTag"` //hàng mới list
	CanTransfer       bool   `json:"canTransfer"`
	Offline           bool   `json:"offline"`
	AlphaID           string `json:"alphaId"`
	Offsell           bool   `json:"offsell"`
	PriceHigh24h      string `gorm:"type:varchar(255)" json:"priceHigh24h"`
	PriceLow24h       string `gorm:"type:varchar(255)" json:"priceLow24h"`
	OnlineTge         bool   `json:"onlineTge"`
	OnlineAirdrop     bool   `json:"onlineAirdrop"`
	CexOffDisplay     bool   `json:"cexOffDisplay"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (AlphaSymbol) TableName() string {
	return "alpha_symbols"
}
