package services

import (
	"chatbtc/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const binanceAPIURL = "https://api.binance.com/api/v3/exchangeInfo"

// FetcherService lấy dữ liệu từ Binance API
type FetcherService struct{}

// NewFetcherService tạo instance mới của service
func NewFetcherService() *FetcherService {
	return &FetcherService{}
}

// FetchAndUpdateSymbols lấy danh sách symbol từ Binance API và cập nhật vào database
func (s *FetcherService) FetchAndUpdateSymbols() error {
	// kiểm tra cập nhật
	if !models.NewSymbolRepository().ShouldUpdate() {
		log.Println("Dữ liệu đã được cập nhật, bỏ qua việc lấy dữ liệu mới")
		return nil
	}

	// lấy dữ liệu mới
	symbols, err := s.fetchFromAPI()
	if err != nil {
		log.Printf("Lỗi khi lấy dữ liệu từ Binance API: %v", err)
		return err
	}

	// lưu dữ liệu vào database
	if err := models.NewSymbolRepository().SaveToDatabase(symbols); err != nil {
		log.Printf("Lỗi khi lưu dữ liệu vào database: %v", err)
	}
	// cập nhật thời gian cập nhật
	if err := models.NewSymbolRepository().UpdateLastUpdateTime(); err != nil {
		log.Printf("Lỗi khi cập nhật thời gian cập nhật: %v", err)
	}
	return nil
}

// fetchFromAPI lấy danh sách symbol từ Binance API
func (s *FetcherService) fetchFromAPI() ([]models.Symbol, error) {
	resp, err := http.Get(binanceAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var exchangeInfo models.BinanceExchangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&exchangeInfo); err != nil {
		return nil, err
	}

	// Lọc symbols với quoteAsset là USDT
	var symbols []models.Symbol
	for _, binanceSymbol := range exchangeInfo.Symbols {
		if binanceSymbol.QuoteAsset == "USDT" && binanceSymbol.Status == "TRADING" {
			symbol := models.Symbol{
				Symbol:    binanceSymbol.Symbol,
				Status:    binanceSymbol.Status,
				BaseAsset: binanceSymbol.BaseAsset,
			}
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

type Scheduler struct {
	fetchService *FetcherService
	stopChan     chan bool
}

func NewScheduler(fetchService *FetcherService) *Scheduler {
	return &Scheduler{
		fetchService: fetchService,
		stopChan:     make(chan bool),
	}
}

func (s *Scheduler) Start() {
	log.Println("Scheduler started")
	// Chạy cập nhật đầu tiên
	go s.runUpdate()

	// Chạy cập nhật định kỳ mỗi 15 ngày
	ticker := time.NewTicker(15 * 24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go s.runUpdate()
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) runUpdate() {
	log.Println("Running update")
	if err := s.fetchService.FetchAndUpdateSymbols(); err != nil {
		log.Printf("Lỗi khi cập nhật dữ liệu: %v", err)
	}
	log.Println("Update completed")
}

func (s *Scheduler) Stop() {
	s.stopChan <- true
}
