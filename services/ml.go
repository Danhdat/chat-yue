package services

import (
	"chatbtc/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MLService struct {
	volumeRepo *models.AutoVolumeRecordRepository
	indicators *TechnicalAnalysisService
}

func NewMLService() *MLService {
	return &MLService{
		volumeRepo: models.NewAutoVolumeRecordRepository(),
		indicators: NewTechnicalAnalysisService(),
	}
}

func (s *AutoVolumeService) fetchRegularKlinesforMLsystem(symbol string) ([][]interface{}, error) {
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

func (s *AutoVolumeService) fetchAlphaKlinesforMLsystem(symbol string) ([][]interface{}, error) {
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
