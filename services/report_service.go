package services

import (
	"chatbtc/models"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ReportService struct {
	volumeRepo          *models.AutoVolumeRecordRepository
	symbolRepo          *models.SymbolRepository
	notificationLogRepo *models.NotificationLogRepository
	alphaRepo           *models.AlphaSymbolRepository
	holderHistoryRepo   *models.HolderHistoryRepository
	telegramBotService  *TelegramBotService
}

func NewReportService(telegramBotService *TelegramBotService) *ReportService {
	return &ReportService{
		volumeRepo:          models.NewAutoVolumeRecordRepository(),
		symbolRepo:          models.NewSymbolRepository(),
		notificationLogRepo: models.NewNotificationLogRepository(),
		alphaRepo:           models.NewAlphaSymbolRepository(),
		holderHistoryRepo:   models.NewHolderHistoryRepository(),
		telegramBotService:  telegramBotService,
	}
}

// Báo cáo top 10 symbol xuất hiện nhiều nhất trong tuần và gửi lên Telegram
func (s *ReportService) ReportTop10SymbolsThisWeek(channelID string) error {
	logs, err := s.notificationLogRepo.GetLogsThisWeek()
	if err != nil {
		return err
	}

	// Tạo struct để lưu thông tin chi tiết
	type symbolInfo struct {
		Count   int
		Bullish int
		Bearish int
	}

	regularSymbols := make(map[string]*symbolInfo)
	alphaSymbols := make(map[string]*symbolInfo)

	// Đếm số lần và phân loại direction
	for _, log := range logs {
		var m map[string]*symbolInfo
		if log.Type == 1 {
			m = alphaSymbols
		} else {
			m = regularSymbols
		}

		if _, exists := m[log.Symbol]; !exists {
			m[log.Symbol] = &symbolInfo{}
		}

		m[log.Symbol].Count++
		switch log.Direction {
		case 1:
			m[log.Symbol].Bullish++
		case 2:
			m[log.Symbol].Bearish++
		}
	}

	// Hàm sắp xếp
	getTop10 := func(m map[string]*symbolInfo) []struct {
		Symbol string
		Info   *symbolInfo
	} {
		var sorted []struct {
			Symbol string
			Info   *symbolInfo
		}
		for k, v := range m {
			sorted = append(sorted, struct {
				Symbol string
				Info   *symbolInfo
			}{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Info.Count > sorted[j].Info.Count
		})
		if len(sorted) > 10 {
			return sorted[:10]
		}
		return sorted
	}

	// Tạo báo cáo
	var builder strings.Builder
	builder.WriteString("📊 *Top 10 Coin tuần này*\n\n")

	// Top Regular
	builder.WriteString("🔹 *Top 10 thường*\n")
	for i, item := range getTop10(regularSymbols) {
		builder.WriteString(fmt.Sprintf("%d. %s: %d lần (↑%d ↓%d)\n",
			i+1,
			item.Symbol,
			item.Info.Count,
			item.Info.Bullish,
			item.Info.Bearish))
	}

	// Top Alpha
	builder.WriteString("\n🔸 *Top 10 Alpha*\n")
	for i, item := range getTop10(alphaSymbols) {
		builder.WriteString(fmt.Sprintf("%d. %s: %d lần (↑%d ↓%d)\n",
			i+1,
			item.Symbol,
			item.Info.Count,
			item.Info.Bullish,
			item.Info.Bearish))
	}

	s.telegramBotService.SendTelegramToChannel(channelID, builder.String())
	return nil
}

type Scheduler4 struct {
	reportService *ReportService
	channelID     string
	stopChan      chan bool
}

func NewScheduler4(reportService *ReportService, channelID string) *Scheduler4 {
	return &Scheduler4{
		reportService: reportService,
		channelID:     channelID,
		stopChan:      make(chan bool),
	}
}

func (s *Scheduler4) Run() {
	if err := s.reportService.ReportTop10SymbolsThisWeek(s.channelID); err != nil {
		logrus.Errorf("Lỗi khi gửi báo cáo: %v", err)
	}
	logrus.Info("Report completed")
}
func (s *Scheduler4) Stop() {
	s.stopChan <- true
}

func (s *Scheduler4) Start() {
	// Thiết lập múi giờ UTC+7
	loc := time.FixedZone("UTC+7", 7*60*60)

	// Hàm helper để tính thời gian đến 10:30 hoặc 18:30 tiếp theo theo múi giờ UTC+7
	nextSchedule := func() time.Time {
		now := time.Now().In(loc) // Chuyển thời gian hiện tại sang UTC+7

		// Tạo thời điểm 10:30 và 18:30 hôm nay theo UTC+7
		today6_30 := time.Date(now.Year(), now.Month(), now.Day(), 6, 30, 0, 0, loc)
		today10_30 := time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, loc)
		today15_30 := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, loc)
		today18_30 := time.Date(now.Year(), now.Month(), now.Day(), 18, 30, 0, 0, loc)
		today21_30 := time.Date(now.Year(), now.Month(), now.Day(), 21, 30, 0, 0, loc)

		// Xác định thời điểm chạy tiếp theo
		switch {
		case now.Before(today6_30):
			return today6_30
		case now.Before(today10_30):
			return today10_30
		case now.Before(today15_30):
			return today15_30
		case now.Before(today18_30):
			return today18_30
		case now.Before(today21_30):
			return today21_30
		default:
			// Nếu đã qua cả 5 mốc, trả về 6:30 ngày hôm sau
			return today6_30.Add(24 * time.Hour)
		}
	}

	// Tạo timer với thời gian đến lần chạy tiếp theo
	timer := time.NewTimer(time.Until(nextSchedule()))
	defer timer.Stop()
	go s.Run()
	for {
		select {
		case <-timer.C:
			go s.Run()
			timer.Reset(time.Until(nextSchedule()))
		case <-s.stopChan:
			logrus.Info("Scheduler stopped")
			return
		}
	}
}

type Scheduler5 struct {
	reportService *ReportService
	stopChan      chan bool
}

func NewScheduler5(reportService *ReportService) *Scheduler5 {
	return &Scheduler5{
		reportService: reportService,
		stopChan:      make(chan bool),
	}
}

func (s *Scheduler5) Run() {
	if err := s.reportService.notificationLogRepo.DeleteLogMonth(); err != nil {
		logrus.Errorf("Lỗi khi xóa log tháng: %v", err)
	}
	logrus.Info("Xóa log tháng thành công")
}
func (s *Scheduler5) Stop() {
	s.stopChan <- true
}
func (s *Scheduler5) Start() {
	// Thiết lập múi giờ UTC+7
	loc := time.FixedZone("UTC+7", 7*60*60)

	// Hàm helper để tính thời gian đến 23:55 ngày 28 tháng tiếp theo
	nextSchedule := func() time.Time {
		now := time.Now().In(loc) // Chuyển thời gian hiện tại sang UTC+7

		// Tạo thời điểm 23:55 ngày 28 tháng hiện tại
		currentMonth28 := time.Date(now.Year(), now.Month(), 28, 23, 55, 0, 0, loc)

		// Nếu đã qua ngày 28 tháng này, tính ngày 28 tháng tiếp theo
		if now.After(currentMonth28) {
			// Chuyển sang tháng tiếp theo
			nextMonth := currentMonth28.AddDate(0, 1, 0)
			return nextMonth
		}

		return currentMonth28
	}

	// Tạo timer với thời gian đến lần chạy tiếp theo
	timer := time.NewTimer(time.Until(nextSchedule()))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			go s.Run()
			timer.Reset(time.Until(nextSchedule()))
		case <-s.stopChan:
			logrus.Info("Scheduler stopped")
			return
		}
	}
}
