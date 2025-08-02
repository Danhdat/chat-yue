package services

import (
	"bytes"
	"chatbtc/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type MLService struct {
	volumeRepo *models.AutoVolumeRecordRepository
	alphaRepo  *models.AlphaSymbolRepository
	indicators *TechnicalAnalysisService
}

// Cấu trúc cho OpenAI API request
type OpenAIRequest struct {
	Prompt struct {
		ID        string                 `json:"id"`
		Version   string                 `json:"version"`
		Variables map[string]interface{} `json:"variables"`
	} `json:"prompt"`
}

// Cấu trúc cho OpenAI API response
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewMLService() *MLService {
	return &MLService{
		volumeRepo: models.NewAutoVolumeRecordRepository(),
		alphaRepo:  models.NewAlphaSymbolRepository(),
		indicators: NewTechnicalAnalysisService(),
	}
}

// Phương thức chính để phân tích symbol với OpenAI
func (s *MLService) AnalyzeSymbolWithOpenAI(originalSymbol string) (string, error) {
	// Bước 1: Xác định loại symbol và resolve
	symbolType := 0

	if s.isAlphaSymbol(originalSymbol) {
		symbolType = 1
		_, err := s.alphaRepo.GetNameByAlphaSymbol(originalSymbol)
		if err != nil {
			return "", fmt.Errorf("lỗi lấy tên symbol cho alpha symbol '%s': %v", originalSymbol, err)
		}
	}

	// Bước 2: Lấy dữ liệu kline
	var klines [][]interface{}
	var err error

	if symbolType == 1 {
		klines, err = s.fetchAlphaKlinesforMLsystem(originalSymbol + "USDT")
	} else {
		klines, err = s.fetchRegularKlinesforMLsystem(originalSymbol)
	}

	if err != nil {
		return "", fmt.Errorf("lỗi lấy dữ liệu %s: %v", originalSymbol, err)
	}

	// Bước 3: Xử lý klines - loại bỏ nến cuối cùng (đang hình thành)
	if len(klines) > 1 {
		klines = klines[:len(klines)-1]
	}

	// Bước 4: Tính toán các indicators
	rsiData, smaData, ema9Data, ema21Data, ema50Data, err := s.calculateIndicators(klines)
	if err != nil {
		return "", fmt.Errorf("lỗi tính toán indicators: %v", err)
	}

	// Bước 5: Chuẩn bị dữ liệu cho OpenAI
	klineData := s.formatKlineData(klines)

	// Bước 6: Gọi OpenAI API
	result, err := s.callOpenAIAPI(originalSymbol, klineData, rsiData, smaData, ema9Data, ema21Data, ema50Data)
	if err != nil {
		return "", fmt.Errorf("lỗi gọi OpenAI API: %v", err)
	}

	return result, nil
}

// Kiểm tra xem symbol có phải là alpha symbol không
func (s *MLService) isAlphaSymbol(symbol string) bool {
	// Logic kiểm tra alpha symbol - có thể cần điều chỉnh theo logic hiện tại
	return strings.HasPrefix(symbol, "ALPHA_")
}

// Tính toán các indicators
func (s *MLService) calculateIndicators(klines [][]interface{}) (string, string, string, string, string, error) {
	// Chuyển đổi klines thành dữ liệu số
	var prices []float64

	for _, kline := range klines {
		closePrice, _ := strconv.ParseFloat(kline[4].(string), 64)
		prices = append(prices, closePrice)
	}

	// Tính RSI (chỉ lấy giá trị cuối cùng)
	rsiValue := s.indicators.CalculateRSI(prices, 14)
	rsiData := fmt.Sprintf("1:%.4f", rsiValue)

	// Tính SMA (tính trung bình của 20 giá cuối)
	var smaData string
	if len(prices) >= 20 {
		smaValue := s.indicators.calculateSMA(prices[len(prices)-20:], 20)
		smaData = fmt.Sprintf("1:%.4f", smaValue)
	} else {
		smaData = "1:0.0000"
	}

	// Tính EMA 9
	ema9Value := s.indicators.CalculateEMA(prices, 9)
	ema9Data := fmt.Sprintf("1:%.4f", ema9Value)

	// Tính EMA 21
	ema21Value := s.indicators.CalculateEMA(prices, 21)
	ema21Data := fmt.Sprintf("1:%.4f", ema21Value)

	// Tính EMA 50
	ema50Value := s.indicators.CalculateEMA(prices, 50)
	ema50Data := fmt.Sprintf("1:%.4f", ema50Value)

	return rsiData, smaData, ema9Data, ema21Data, ema50Data, nil
}

// Format dữ liệu kline
func (s *MLService) formatKlineData(klines [][]interface{}) string {
	var result []string
	for i, kline := range klines {
		// Format: "index:open,high,low,close,volume"
		line := fmt.Sprintf("%d:%s,%s,%s,%s,%s",
			i+1,
			kline[1], // open
			kline[2], // high
			kline[3], // low
			kline[4], // close
			kline[7], // quote asset volume
		)
		result = append(result, line)
	}
	return strings.Join(result, ";")
}

// Format dữ liệu indicator
func (s *MLService) formatIndicatorData(values []float64) string {
	var result []string
	for i, value := range values {
		result = append(result, fmt.Sprintf("%d:%.4f", i+1, value))
	}
	return strings.Join(result, ";")
}

// Gọi OpenAI API
func (s *MLService) callOpenAIAPI(symbol, klineData, rsiData, smaData, ema9Data, ema21Data, ema50Data string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY không được thiết lập")
	}

	request := OpenAIRequest{}
	request.Prompt.ID = "pmpt_688998fee2bc819491c6112cafd3cea1070efa35b2150e15"
	request.Prompt.Version = "9"
	request.Prompt.Variables = map[string]interface{}{
		"symbol":     symbol,
		"kline_data": klineData,
		"rsi_data":   rsiData,
		"sma_data":   smaData,
		"ema9_data":  ema9Data,
		"ema21_data": ema21Data,
		"ema50_data": ema50Data,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("lỗi marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("lỗi tạo request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lỗi gọi API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("lỗi đọc response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API trả về lỗi %d: %s", resp.StatusCode, string(body))
	}

	var openAIResponse OpenAIResponse
	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		return "", fmt.Errorf("lỗi parse response: %v", err)
	}

	if len(openAIResponse.Choices) == 0 {
		return "", fmt.Errorf("không có response từ OpenAI")
	}

	return openAIResponse.Choices[0].Message.Content, nil
}

func (s *MLService) fetchRegularKlinesforMLsystem(symbol string) ([][]interface{}, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1d&limit=1000", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, fmt.Errorf("lỗi parse regular klines: %w", err)
	}

	return klines, nil
}

func (s *MLService) fetchAlphaKlinesforMLsystem(symbol string) ([][]interface{}, error) {
	url := fmt.Sprintf("https://www.binance.com/bapi/defi/v1/public/alpha-trade/klines?interval=1d&limit=1000&symbol=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cấu trúc response mới của API Alpha
	var alphaResponse struct {
		Code    string          `json:"code"`
		Message interface{}     `json:"message"`
		Data    [][]interface{} `json:"data"`
		Success bool            `json:"success"`
	}

	if err := json.Unmarshal(body, &alphaResponse); err != nil {
		return nil, fmt.Errorf("lỗi parse alpha klines: %w", err)
	}

	if !alphaResponse.Success || alphaResponse.Code != "000000" {
		return nil, fmt.Errorf("API Alpha trả về lỗi: %v", alphaResponse.Message)
	}

	return alphaResponse.Data, nil
}
