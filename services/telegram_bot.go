package services

import (
	"chatbtc/config"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"chatbtc/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
)

// TelegramBotService quản lý bot Telegram
type TelegramBotService struct {
	bot        *tgbotapi.BotAPI
	cryptoAPI  *CryptoAPIService
	indicators *TechnicalAnalysisService
	analysis   *AnalysisService
	mlService  *MLService
	chatID     int64
	channelID  string
}

// NewTelegramBotService tạo instance mới của service
func NewTelegramBotService() (*TelegramBotService, error) {
	// Cấu hình proxy nếu được bật
	var client *http.Client
	if config.AppConfig.ProxyEnabled && config.AppConfig.ProxyURL != "" {
		switch config.AppConfig.ProxyType {
		case "socks5":
			// Xử lý SOCKS5
			auth := &proxy.Auth{
				User:     config.AppConfig.ProxyUsername,
				Password: config.AppConfig.ProxyPassword,
			}
			dialer, err := proxy.SOCKS5("tcp", strings.TrimPrefix(config.AppConfig.ProxyURL, "socks5://"), auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("lỗi kết nối SOCKS5 proxy: %v", err)
			}
			client = &http.Client{
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
				Timeout: 30 * time.Second,
			}
		default:
			// Xử lý HTTP/HTTPS
			proxyURL, err := url.Parse(config.AppConfig.ProxyURL)
			if err != nil {
				return nil, fmt.Errorf("lỗi parse proxy URL: %v", err)
			}
			if config.AppConfig.ProxyUsername != "" {
				proxyURL.User = url.UserPassword(config.AppConfig.ProxyUsername, config.AppConfig.ProxyPassword)
			}
			client = &http.Client{
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(proxyURL),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Bỏ qua SSL nếu cần
				},
				Timeout: 30 * time.Second,
			}
		}
	} else {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// Tạo bot thông thường trước
	bot, err := tgbotapi.NewBotAPI(config.AppConfig.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("lỗi khởi tạo bot: %v", err)
	}
	if config.AppConfig.ProxyEnabled {
		bot.Client = client
	}

	// Chat ID không bắt buộc, bot sẽ nhận tin nhắn từ tất cả users
	var chatID int64
	if config.AppConfig.TelegramChatID != "" {
		if parsedChatID, err := strconv.ParseInt(config.AppConfig.TelegramChatID, 10, 64); err == nil {
			chatID = parsedChatID
		}
	}
	channelID := "@yuealerts"
	return &TelegramBotService{
		bot:        bot,
		cryptoAPI:  NewCryptoAPIService(),
		indicators: NewTechnicalAnalysisService(),
		analysis:   NewAnalysisService(),
		mlService:  NewMLService(),
		chatID:     chatID,
		channelID:  channelID,
	}, nil
}

// StartBot khởi động bot
func (s *TelegramBotService) StartBot() {
	log.Println("🚀 Khởi động Crypto Analysis Bot...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)
	log.Println("✅ Bot đã sẵn sàng nhận tin nhắn...")
	log.Printf("Bot đã khởi động: %s", s.bot.Self.UserName)

	for update := range updates {
		// Xử lý callback query (button clicks)
		if update.CallbackQuery != nil {
			s.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		// Xử lý tin nhắn thông thường
		if update.Message == nil {
			continue
		}

		// Log tin nhắn nhận được
		log.Printf("Nhận tin nhắn từ %s , %s: %s", update.Message.From.UserName, update.Message.Chat.Title, update.Message.Text)

		// Xử lý tin nhắn
		s.handleMessage(update.Message)
	}
}

// handleMessage xử lý tin nhắn từ user
func (s *TelegramBotService) handleMessage(message *tgbotapi.Message) {
	text := message.Text
	if !strings.HasPrefix(text, "/") {
		return
	}

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	chatID := message.Chat.ID

	switch command {
	case "/start":
		s.sendWelcomeMessage(chatID)
	case "/help":
		s.sendHelpMessage(chatID)
	case "/price":
		if len(parts) < 2 {
			s.sendMessage(chatID, "❌ Vui lòng cung cấp symbol. Ví dụ: /price BTCUSDT")
			return
		}
		symbol := strings.ToUpper(parts[1])
		s.handlePriceCommand(chatID, symbol)
	case "/analyze":
		if len(parts) < 3 {
			s.sendMessage(chatID, "❌ Vui lòng cung cấp interval và symbol. Ví dụ: /analyze 1h BTCUSDT")
			return
		}
		interval := parts[1]
		symbol := strings.ToUpper(parts[2])
		s.handleAnalyzeCommand(chatID, interval, symbol)
	default:
		s.sendMessage(chatID, "❌ Lệnh không hợp lệ. Gõ /help để xem danh sách lệnh.")
	}
}

// handlePriceCommand xử lý lệnh /price
func (s *TelegramBotService) handlePriceCommand(chatID int64, symbol string) {
	price, err := s.cryptoAPI.GetCurrentPrice(symbol)
	if err != nil {
		s.sendMessage(chatID, fmt.Sprintf("❌ Lỗi khi lấy giá %s: %v", symbol, err))
		return
	}

	message := fmt.Sprintf("💰 **Giá %s**\n\n", strings.ToUpper(symbol))
	message += fmt.Sprintf("💵 **Giá hiện tại:** $%s\n", utils.FormatPrice(price.CurrentPrice))
	message += fmt.Sprintf("📈 **Thay đổi 24h:** %s (%s)\n",
		utils.FormatPercentage(price.PriceChangePercentage24h), utils.FormatPrice(price.PriceChange24h))
	message += fmt.Sprintf("📊 **Volume 24h:** $%s\n", utils.FormatVolume(price.Volume24h))
	loc := time.FixedZone("UTC+7", 7*60*60)
	message += fmt.Sprintf("⏰ **Cập nhật:** %s", price.LastUpdated.In(loc).Format("15:04:05 02/01/2006"))

	s.sendMessage(chatID, message)
}

// handleAnalyzeCommand xử lý lệnh /analyze
func (s *TelegramBotService) handleAnalyzeCommand(chatID int64, interval string, symbol string) {
	// Validate interval
	validIntervals := map[string]bool{
		"1m": true, "5m": true, "15m": true, "30m": true,
		"1h": true, "4h": true, "1d": true, "1w": true,
	}

	if !validIntervals[interval] {
		s.sendMessage(chatID, "❌ Interval không hợp lệ. Các interval được hỗ trợ: 1m, 5m, 15m, 30m, 1h, 4h, 1d, 1w")
		return
	}

	log.Printf("Bắt đầu phân tích symbol: %s với interval: %s", symbol, interval)

	// Lấy dữ liệu kline (100 nến gần nhất)
	klines, err := s.cryptoAPI.GetKlineData(symbol, interval, 100)
	if err != nil {
		s.sendMessage(chatID, fmt.Sprintf("❌ Lỗi khi lấy dữ liệu %s: %v", symbol, err))
		return
	}

	log.Printf("Đã lấy được %d điểm dữ liệu lịch sử với interval %s", len(klines), interval)

	// Phân tích với service indicators mới
	analysis, err := s.indicators.AnalyzeCrypto(symbol, klines, interval)
	if err != nil {
		s.sendMessage(chatID, fmt.Sprintf("❌ Lỗi khi phân tích %s: %v", symbol, err))
		return
	}

	log.Printf("Gửi tin nhắn phân tích cho chatID: %d", chatID)
	s.sendMessage(chatID, analysis)

	// Lấy dữ liệu phân tích chi tiết để lưu vào database
	analysisData, err := s.indicators.GetAnalysisData(symbol, klines, interval)
	if err != nil {
		log.Printf("⚠️ Lỗi lấy dữ liệu phân tích: %v", err)
		return
	}

	// Lưu kết quả phân tích vào database
	if len(klines) > 0 {
		latestKline := klines[len(klines)-1]

		// Chuyển đổi kiểu dữ liệu
		closePrice, _ := strconv.ParseFloat(latestKline.Close, 64)
		volume, _ := strconv.ParseFloat(latestKline.QuoteAssetVolume, 64)

		err := s.analysis.SaveAnalysis(
			symbol,
			interval,
			closePrice,
			volume,
			analysisData.RSI,
			analysisData.EMA9,
			analysisData.EMA21,
			analysisData.EMA50,
			analysisData.MACD,
			analysisData.MACDSignal,
			analysisData.VolumeSMA,
			analysisData.Trend,
			analysisData.Power,
			analysisData.Signal,
			analysisData.Recommendation,
			analysisData.VolumeSignal,
		)
		if err != nil {
			log.Printf("⚠️ Lỗi lưu phân tích: %v", err)
		} else {
			log.Printf("✅ Đã lưu phân tích cho %s (%s)", symbol, interval)
		}
	}
}

// sendWelcomeMessage gửi tin nhắn chào mừng
func (s *TelegramBotService) sendWelcomeMessage(chatID int64) {
	message := "🎉 **Chào mừng đến với Crypto Analysis Bot!**\n\n"
	message += "🤖 Tôi có thể giúp bạn:\n"
	message += "• Xem giá crypto hiện tại\n"
	message += "• Phân tích kỹ thuật với RSI và EMA\n"
	message += "• Đưa ra khuyến nghị giao dịch\n\n"
	message += "📝 **Các lệnh có sẵn:**\n"
	message += "/price <symbol> - Xem giá hiện tại\n"
	message += "/analyze <interval> <symbol> - Phân tích kỹ thuật\n"
	message += "/help - Xem hướng dẫn chi tiết\n\n"
	message += "💡 **Ví dụ:**\n"
	message += "/price BTCUSDT\n"
	message += "/analyze 1h BTCUSDT"

	s.sendMessage(chatID, message)
}

// sendHelpMessage gửi tin nhắn hướng dẫn
func (s *TelegramBotService) sendHelpMessage(chatID int64) {
	message := "📚 **Hướng dẫn sử dụng Crypto Analysis Bot**\n\n"
	message += "🔹 **Lệnh cơ bản:**\n"
	message += "• `/start` - Khởi động bot\n"
	message += "• `/help` - Xem hướng dẫn này\n\n"
	message += "🔹 **Lệnh phân tích:**\n"
	message += "• `/price <symbol>` - Xem giá hiện tại\n"
	message += "• `/analyze <interval> <symbol>` - Phân tích kỹ thuật\n\n"
	message += "🔹 **Interval được hỗ trợ:**\n"
	message += "• `1m` - 1 phút\n"
	message += "• `5m` - 5 phút\n"
	message += "• `15m` - 15 phút\n"
	message += "• `30m` - 30 phút\n"
	message += "• `1h` - 1 giờ\n"
	message += "• `4h` - 4 giờ\n"
	message += "• `1d` - 1 ngày\n"
	message += "• `1w` - 1 tuần\n\n"
	message += "🔹 **Symbol được hỗ trợ:**\n"
	message += "• BTCUSDT, ETHUSDT, ADAUSDT\n"
	message += "• BNBUSDT, DOTUSDT, LINKUSDT\n"
	message += "• Và nhiều cặp khác...\n\n"
	message += "🔹 **Chỉ báo kỹ thuật:**\n"
	message += "• RSI (14) - Chỉ báo quá mua/quá bán\n"
	message += "• EMA (20) - Đường trung bình động\n\n"
	message += "💡 **Ví dụ sử dụng:**\n"
	message += "• `/analyze 1h BTCUSDT` - Phân tích BTC theo nến 1h\n"
	message += "• `/analyze 15m ETHUSDT` - Phân tích ETH theo nến 15m\n"
	message += "• `/analyze 1d BNBUSDT` - Phân tích BNB theo nến 1 ngày\n\n"
	message += "⚠️ **Lưu ý:**\n"
	message += "• Chỉ mang tính chất tham khảo\n"
	message += "• Không phải lời khuyên đầu tư\n"
	message += "• Luôn DYOR (Do Your Own Research)"

	s.sendMessage(chatID, message)
}

// sendMessage gửi tin nhắn đến chat
func (s *TelegramBotService) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err := s.bot.Send(msg)
	if err != nil {
		log.Printf("Lỗi khi gửi tin nhắn: %v", err)
	}
}

// sendMessage gửi tin nhắn đến channel

func (s *TelegramBotService) SendTelegramToChannel(channelID, message string) {
	log.Println("Sending message to channel: ", channelID, message)
	msg := tgbotapi.NewMessageToChannel(channelID, message)
	msg.ParseMode = "Markdown"
	_, err := s.bot.Send(msg)
	if err != nil {
		log.Printf("Lỗi khi gửi tin nhắn đến channel: %v", err)
	}
}

// SendTelegramToChannelWithAdvancedButton gửi tin nhắn đến channel với button phân tích nâng cao
func (s *TelegramBotService) SendTelegramToChannelWithAdvancedButton(channelID, message string, symbol string) {
	log.Println("Sending message with advanced analysis button to channel: ", channelID)

	// Thực hiện phân tích OpenAI trước
	openAIAnalysis, err := s.mlService.AnalyzeSymbolWithOpenAI(symbol)
	if err != nil {
		log.Printf("Lỗi khi phân tích OpenAI cho symbol %s: %v", symbol, err)
		// Nếu lỗi, vẫn gửi tin nhắn gốc nhưng không có phân tích OpenAI
		openAIAnalysis = "❌ Không thể phân tích nâng cao do lỗi kỹ thuật"
	}

	// Thêm phân tích OpenAI vào message
	enhancedMessage := message + "\n\n" + "🤖 **Phân tích AI:**\n" + openAIAnalysis

	msg := tgbotapi.NewMessageToChannel(channelID, enhancedMessage)
	msg.ParseMode = "Markdown"

	// Tạo inline keyboard với button phân tích nâng cao
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Cập nhật phân tích", fmt.Sprintf("advanced_analysis_%s", symbol)),
		),
	)

	msg.ReplyMarkup = keyboard

	_, err = s.bot.Send(msg)
	if err != nil {
		log.Printf("Lỗi khi gửi tin nhắn với button đến channel: %v", err)
	}
}

func (s *TelegramBotService) GetChannelID() string {
	return s.channelID
}

// Stop dừng bot
func (s *TelegramBotService) Stop() {
	log.Println("🛑 Đang dừng bot...")
}

// handleCallbackQuery xử lý callback query từ inline keyboard
func (s *TelegramBotService) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	log.Printf("Nhận callback query: %s", callback.Data)

	// Trả lời callback query để bỏ loading state
	callbackResponse := tgbotapi.NewCallback(callback.ID, "")
	_, err := s.bot.Request(callbackResponse)
	if err != nil {
		log.Printf("Lỗi khi trả lời callback query: %v", err)
	}

	// Xử lý các loại callback khác nhau
	if strings.HasPrefix(callback.Data, "advanced_analysis_") {
		s.handleAdvancedAnalysisCallback(callback)
	}
}

// handleAdvancedAnalysisCallback xử lý callback cho phân tích nâng cao
func (s *TelegramBotService) handleAdvancedAnalysisCallback(callback *tgbotapi.CallbackQuery) {
	// Trích xuất symbol từ callback data
	parts := strings.Split(callback.Data, "_")
	if len(parts) < 3 {
		log.Printf("Callback data không hợp lệ: %s", callback.Data)
		return
	}

	symbol := strings.Join(parts[2:], "_") // Ghép lại các phần còn lại để xử lý symbol có dấu gạch dưới

	log.Printf("Thực hiện phân tích nâng cao cho symbol: %s", symbol)

	// Thực hiện phân tích OpenAI
	openAIAnalysis, err := s.mlService.AnalyzeSymbolWithOpenAI(symbol)
	if err != nil {
		log.Printf("Lỗi khi phân tích OpenAI cho symbol %s: %v", symbol, err)
		openAIAnalysis = "❌ Không thể phân tích nâng cao do lỗi kỹ thuật"
	}

	// Tạo tin nhắn cập nhật
	updateMessage := fmt.Sprintf("🤖 **Phân tích AI cập nhật cho %s:**\n\n%s", symbol, openAIAnalysis)

	// Gửi tin nhắn cập nhật
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, updateMessage)
	msg.ParseMode = "Markdown"
	msg.ReplyToMessageID = callback.Message.MessageID

	_, err = s.bot.Send(msg)
	if err != nil {
		log.Printf("Lỗi khi gửi tin nhắn cập nhật: %v", err)
	}
}
