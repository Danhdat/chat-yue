package services

import (
	"chatbtc/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const binanceAPIURL = "https://api.binance.com/api/v3/exchangeInfo"
const alphaAPIURL = "https://www.binance.com/bapi/defi/v1/public/wallet-direct/buw/wallet/cex/alpha/all/token/list"

// FetcherService lấy dữ liệu từ Binance API
type FetcherService struct{}

// NewFetcherService tạo instance mới của service
func NewFetcherService() *FetcherService {
	return &FetcherService{}
}

// FetchAndUpdateSymbols lấy danh sách symbol từ Binance API và cập nhật vào database
func (s *FetcherService) FetchAndUpdateSymbols() error {
	// kiểm tra cập nhật
	if !models.NewCommonRepository().ShouldUpdate("symbols") {
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
	if err := models.NewCommonRepository().UpdateLastUpdateTime("symbols"); err != nil {
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

// ALpha binance
func (s *FetcherService) fetchAlphaFromAPI() ([]models.AlphaSymbol, error) {
	resp, err := http.Get(alphaAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse models.AlphaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	var symbols []models.AlphaSymbol
	for _, alphaSymbol := range apiResponse.Data {
		symbol := models.AlphaSymbol{
			TokenID:           alphaSymbol.TokenID,
			ChainID:           alphaSymbol.ChainID,
			ChainName:         alphaSymbol.ChainName,
			ContractAddress:   alphaSymbol.ContractAddress,
			Symbol:            alphaSymbol.Symbol,
			PercentChange24h:  alphaSymbol.PercentChange24h,
			Volume24h:         alphaSymbol.Volume24h,
			MarketCap:         alphaSymbol.MarketCap,
			FDV:               alphaSymbol.FDV,
			Liquidity:         alphaSymbol.Liquidity,
			TotalSupply:       alphaSymbol.TotalSupply,
			CirculatingSupply: alphaSymbol.CirculatingSupply,
			Holders:           alphaSymbol.Holders,
			ListingCex:        alphaSymbol.ListingCex,
			HotTag:            alphaSymbol.HotTag,
			CanTransfer:       alphaSymbol.CanTransfer,
			Offline:           alphaSymbol.Offline,
			AlphaID:           alphaSymbol.AlphaID,
			Offsell:           alphaSymbol.Offsell,
			PriceHigh24h:      alphaSymbol.PriceHigh24h,
			PriceLow24h:       alphaSymbol.PriceLow24h,
			OnlineTge:         alphaSymbol.OnlineTge,
			OnlineAirdrop:     alphaSymbol.OnlineAirdrop,
			CexOffDisplay:     alphaSymbol.CexOffDisplay,
		}
		symbols = append(symbols, symbol)
	}
	return symbols, nil
}

func (s *FetcherService) FetchAndUpdateAlpha() error {
	// kiểm tra cập nhật
	if !models.NewCommonRepository().ShouldUpdate("alpha_symbols") {
		log.Println("Dữ liệu Alpha đã được cập nhật, bỏ qua việc lấy dữ liệu mới")
		return nil
	}

	// lấy dữ liệu mới
	symbols, err := s.fetchAlphaFromAPI()
	if err != nil {
		log.Printf("Lỗi khi lấy dữ liệu từ Binance API Alpha: %v", err)
		return err
	}

	// lưu dữ liệu vào database
	if err := models.NewAlphaSymbolRepository().SaveToDatabaseAlpha(symbols); err != nil {
		log.Printf("Lỗi khi lưu dữ liệu Alpha vào database: %v", err)
		return err
	}

	// Lưu snapshot holders vào holder_history
	if err := s.saveHoldersSnapshot(symbols); err != nil {
		log.Printf("Lỗi khi lưu snapshot holders: %v", err)
	}

	// cập nhật thời gian cập nhật
	if err := models.NewCommonRepository().UpdateLastUpdateTime("alpha_symbols"); err != nil {
		log.Printf("Lỗi khi cập nhật thời gian cập nhật Alpha: %v", err)
	}
	return nil
}

// saveHoldersSnapshot lưu snapshot holders vào bảng holder_history
func (s *FetcherService) saveHoldersSnapshot(symbols []models.AlphaSymbol) error {
	holderRepo := models.NewHolderHistoryRepository()

	for _, symbol := range symbols {
		// Debug: Log dữ liệu raw từ API
		log.Printf("DEBUG: Symbol %s - Raw Holders: '%s'", symbol.Symbol, symbol.Holders)

		// Convert Holders từ string sang int
		holdersCount := 0
		if symbol.Holders != "" {
			// Loại bỏ các ký tự không phải số
			cleanHolders := strings.TrimSpace(symbol.Holders)
			// Loại bỏ dấu phẩy, chấm nếu có
			cleanHolders = strings.ReplaceAll(cleanHolders, ",", "")
			cleanHolders = strings.ReplaceAll(cleanHolders, ".", "")

			// Thử parse với strconv.Atoi
			if count, err := strconv.Atoi(cleanHolders); err == nil {
				holdersCount = count
			} else {
				// Nếu không được, thử với fmt.Sscanf
				if _, err := fmt.Sscanf(cleanHolders, "%d", &holdersCount); err != nil {
					log.Printf("Lỗi convert holders cho symbol %s: '%s' (cleaned: '%s') - %v",
						symbol.Symbol, symbol.Holders, cleanHolders, err)
					continue
				}
			}
		}

		// Debug: Log số đã convert
		log.Printf("DEBUG: Symbol %s - Converted Holders: %d", symbol.Symbol, holdersCount)

		// Kiểm tra dữ liệu hợp lý (holders không quá lớn)
		if holdersCount > 1000000 {
			log.Printf("WARNING: Symbol %s có holders quá lớn (%d), có thể lỗi dữ liệu", symbol.Symbol, holdersCount)
			continue
		}

		// Lấy record cũ nhất để tính change_amount
		latestHistory, err := holderRepo.GetLatestBySymbol(symbol.Symbol)
		changeAmount := 0.0

		if err == nil && latestHistory != nil {
			// Có dữ liệu cũ, tính thay đổi
			changeAmount = float64(holdersCount - latestHistory.Holders)
			log.Printf("DEBUG: Symbol %s - Previous: %d, Current: %d, Change: %.0f",
				symbol.Symbol, latestHistory.Holders, holdersCount, changeAmount)
		} else {
			log.Printf("DEBUG: Symbol %s - First time, no previous data", symbol.Symbol)
		}

		// Tạo record mới
		history := &models.HolderHistory{
			Symbol:       symbol.Symbol,
			Holders:      holdersCount,
			ChangeAmount: changeAmount,
		}

		// Lưu vào database
		if err := holderRepo.Create(history); err != nil {
			log.Printf("Lỗi lưu holder history cho symbol %s: %v", symbol.Symbol, err)
			continue
		}
	}

	log.Printf("Đã lưu snapshot holders cho %d symbols", len(symbols))
	return nil
}

type SchedulerAlpha struct {
	fetchService *FetcherService
	stopChan     chan bool
}

func NewSchedulerAlpha(fetchService *FetcherService) *SchedulerAlpha {
	return &SchedulerAlpha{
		fetchService: fetchService,
		stopChan:     make(chan bool),
	}
}
func (s *SchedulerAlpha) Stop() {
	s.stopChan <- true
}
func (s *SchedulerAlpha) Run() {
	log.Println("Running Alpha update")
	if err := s.fetchService.FetchAndUpdateAlpha(); err != nil {
		log.Printf("Lỗi khi cập nhật dữ liệu: %v", err)
	}
	log.Println("Update completed")
}
func (s *SchedulerAlpha) Start() {
	// Hàm helper để tính thời gian đến giờ:30 tiếp theo (cách nhau 2 tiếng)
	nextSchedule := func() time.Time {
		now := time.Now()
		// Tính giờ tiếp theo (làm tròn xuống giờ chẵn rồi cộng 2h30)
		currentHour := now.Truncate(time.Hour) // Làm tròn xuống giờ (VD: 9:45 → 9:00)
		// Thử cộng 2h30 vào giờ hiện tại
		next := currentHour.Add(2*time.Hour + 30*time.Minute)
		// Nếu thời gian này vẫn trong quá khứ (VD: now = 9:45, next = 11:30 → hợp lệ)
		// Nhưng nếu now = 11:45, next = 13:30 → cần kiểm tra
		if next.Before(now) {
			// Nếu đã qua mốc, tính tiếp mốc sau (VD: 13:30 → 15:30)
			next = next.Add(2 * time.Hour)
		}

		return next
	}

	// Tạo timer với thời gian đến lần chạy tiếp theo (VD: 9:30, 11:30,...)
	ticker := time.NewTimer(time.Until(nextSchedule()))
	defer ticker.Stop()
	go s.Run()
	for {
		select {
		case <-ticker.C:
			go s.Run()
			ticker.Reset(time.Until(nextSchedule()))
		case <-s.stopChan:
			log.Println("Scheduler Alpha stopped")
			return
		}
	}
}
