package services

import (
	"chatbtc/models"
)

type MLService struct {
	symbolRepo *models.SymbolRepository
	volumeRepo *models.AutoVolumeRecordRepository
}

func NewMLService() *MLService {
	return &MLService{
		symbolRepo: models.NewSymbolRepository(),
		volumeRepo: models.NewAutoVolumeRecordRepository(),
	}
}
