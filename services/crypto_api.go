package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"chatbtc/models"

	"github.com/shopspring/decimal"
)

// CryptoAPIService cung cấp các phương thức để lấy dữ liệu crypto từ Binance
type CryptoAPIService struct {
	apiURL string
	client *http.Client
}

// NewCryptoAPIService tạo instance mới của service
func NewCryptoAPIService() *CryptoAPIService {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &CryptoAPIService{
		apiURL: "https://api.binance.com/api/v3",
		client: client,
	}
}

// GetKlineData lấy dữ liệu kline từ Binance
func (s *CryptoAPIService) GetKlineData(symbol string, interval string, limit int) ([]models.KlineData, error) {
	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s&limit=%d", s.apiURL, strings.ToUpper(symbol), interval, limit)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi gọi API Binance: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi đọc response: %v", err)
	}

	// Kiểm tra xem response có phải là error object không
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		if code, exists := errorResponse["code"]; exists && code != nil {
			return nil, fmt.Errorf("lỗi API Binance: %v", errorResponse["msg"])
		}
	}

	var rawData [][]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("lỗi khi parse JSON: %v", err)
	}

	var klines []models.KlineData
	for _, data := range rawData {
		if len(data) < 12 {
			continue
		}

		kline := models.KlineData{
			OpenTime:                 int64(data[0].(float64)),
			Open:                     data[1].(string),
			High:                     data[2].(string),
			Low:                      data[3].(string),
			Close:                    data[4].(string),
			Volume:                   data[5].(string),
			CloseTime:                int64(data[6].(float64)),
			QuoteAssetVolume:         data[7].(string),
			NumberOfTrades:           int(data[8].(float64)),
			TakerBuyBaseAssetVolume:  data[9].(string),
			TakerBuyQuoteAssetVolume: data[10].(string),
			Ignore:                   data[11].(string),
		}
		klines = append(klines, kline)
	}

	return klines, nil
}

// GetCurrentPrice lấy thông tin giá và volume 24h từ Binance ticker API
func (s *CryptoAPIService) GetCurrentPrice(symbol string) (*models.CryptoPrice, error) {
	url := fmt.Sprintf("%s/ticker/24hr?symbol=%s", s.apiURL, strings.ToUpper(symbol))

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi gọi API Binance ticker: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi đọc response: %v", err)
	}

	// Kiểm tra xem response có phải là error object không
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		if code, exists := errorResponse["code"]; exists && code != nil {
			return nil, fmt.Errorf("lỗi API Binance: %v", errorResponse["msg"])
		}
	}

	// Parse response từ ticker API
	var tickerData map[string]interface{}
	if err := json.Unmarshal(body, &tickerData); err != nil {
		return nil, fmt.Errorf("lỗi khi parse JSON ticker: %v", err)
	}

	// Chuyển đổi các giá trị sang decimal
	lastPriceStr := getString(tickerData, "lastPrice")
	currentPrice, err := decimal.NewFromString(lastPriceStr)
	if err != nil {
		currentPrice = decimal.Zero
	}

	volume24h, _ := decimal.NewFromString(getString(tickerData, "quoteVolume"))
	priceChange, _ := decimal.NewFromString(getString(tickerData, "priceChange"))
	priceChangePercent, _ := decimal.NewFromString(getString(tickerData, "priceChangePercent"))

	// Chuyển đổi timestamp
	closeTime := int64(tickerData["closeTime"].(float64))

	price := &models.CryptoPrice{
		ID:                       0, // ID sẽ được set sau khi có database
		Symbol:                   getString(tickerData, "symbol"),
		Name:                     getString(tickerData, "symbol"),
		CurrentPrice:             currentPrice,
		MarketCap:                decimal.Zero, // Ticker API không cung cấp market cap
		Volume24h:                volume24h.Round(0),
		PriceChange24h:           priceChange,
		PriceChangePercentage24h: priceChangePercent,
		LastUpdated:              time.Unix(closeTime/1000, 0),
	}

	return price, nil
}

// GetHistoricalData lấy dữ liệu lịch sử giá từ Binance
func (s *CryptoAPIService) GetHistoricalData(symbol string, limit int) ([]models.PriceData, error) {
	klines, err := s.GetKlineData(symbol, "1h", limit)
	if err != nil {
		return nil, err
	}

	var priceData []models.PriceData
	for _, kline := range klines {
		closePrice, _ := decimal.NewFromString(kline.Close)
		volume, _ := decimal.NewFromString(kline.Volume)

		priceData = append(priceData, models.PriceData{
			Timestamp: time.Unix(kline.CloseTime/1000, 0),
			Price:     closePrice,
			Volume:    volume,
		})
	}

	return priceData, nil
}

// helper function để lấy string từ map
func getString(data map[string]interface{}, key string) string {
	if val, exists := data[key]; !exists {
		return ""
	} else {
		// Xử lý nhiều kiểu dữ liệu
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%.8f", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case int:
			return fmt.Sprintf("%d", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
}
