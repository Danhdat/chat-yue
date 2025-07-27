package services

import (
	"chatbtc/models"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type ReportService struct {
	volumeRepo          *models.AutoVolumeRecordRepository
	symbolRepo          *models.SymbolRepository
	notificationLogRepo *models.NotificationLogRepository
	telegramBotService  *TelegramBotService
}

func NewReportService(telegramBotService *TelegramBotService) *ReportService {
	return &ReportService{
		volumeRepo:          models.NewAutoVolumeRecordRepository(),
		symbolRepo:          models.NewSymbolRepository(),
		notificationLogRepo: models.NewNotificationLogRepository(),
		telegramBotService:  telegramBotService,
	}
}

// Báo cáo top 10 symbol xuất hiện nhiều nhất trong tuần và gửi lên Telegram
func (s *ReportService) ReportTop10SymbolsThisWeek(channelID string) error {
	logs, err := s.notificationLogRepo.GetLogsThisWeek()
	if err != nil {
		return err
	}

	symbolCount := make(map[string]int)
	for _, log := range logs {
		symbolCount[log.Symbol]++
	}

	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range symbolCount {
		sorted = append(sorted, kv{k, v})
	}
	// Sắp xếp giảm dần theo số lần xuất hiện
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	top := 10
	if len(sorted) < 10 {
		top = len(sorted)
	}

	var builder strings.Builder
	builder.WriteString("📊 *Top 10 Coin xuất hiện nhiều nhất trong tuần này*\n")
	for i := 0; i < top; i++ {
		builder.WriteString(fmt.Sprintf("%d. %s: %d lần\n", i+1, sorted[i].Key, sorted[i].Value))
	}

	message := builder.String()
	s.telegramBotService.SendTelegramToChannel(channelID, message)
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
		log.Printf("Lỗi khi gửi báo cáo: %v", err)
	}
	log.Println("Report completed")
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
		today10_30 := time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, loc)
		today18_30 := time.Date(now.Year(), now.Month(), now.Day(), 18, 30, 0, 0, loc)

		// Xác định thời điểm chạy tiếp theo
		switch {
		case now.Before(today10_30):
			return today10_30
		case now.Before(today18_30):
			return today18_30
		default:
			// Nếu đã qua cả 2 mốc, trả về 10:30 ngày hôm sau
			return today10_30.Add(24 * time.Hour)
		}
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
			log.Println("Scheduler stopped")
			return
		}
	}
}
