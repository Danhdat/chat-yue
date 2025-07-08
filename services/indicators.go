package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"chatbtc/models"
	"chatbtc/utils"

	"github.com/shopspring/decimal"
)

// TechnicalAnalysisService cung cấp các phương thức tính toán chỉ báo kỹ thuật
type TechnicalAnalysisService struct{}

// NewTechnicalAnalysisService tạo instance mới của service
func NewTechnicalAnalysisService() *TechnicalAnalysisService {
	return &TechnicalAnalysisService{}
}

// CalculateRSI tính toán chỉ báo RSI (Relative Strength Index)
func (s *TechnicalAnalysisService) CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	gains := make([]float64, len(prices)-1)
	losses := make([]float64, len(prices)-1)

	// Tính gain và loss
	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i-1] = change
			losses[i-1] = 0
		} else {
			gains[i-1] = 0
			losses[i-1] = -change
		}
	}

	// Tính trung bình gain và loss
	avgGain := s.calculateSMA(gains[len(gains)-period:], period)
	avgLoss := s.calculateSMA(losses[len(losses)-period:], period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// CalculateEMA tính toán Exponential Moving Average
func (s *TechnicalAnalysisService) CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Multiplier
	multiplier := 2.0 / float64(period+1)

	// Tính SMA đầu tiên làm EMA ban đầu
	ema := s.calculateSMA(prices[:period], period)

	// Tính EMA cho các giá trị còn lại
	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// calculateSMA tính Simple Moving Average
func (s *TechnicalAnalysisService) calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for _, price := range prices {
		sum += price
	}

	return sum / float64(len(prices))
}

// CalculateMACD tính toán MACD (Moving Average Convergence Divergence)
func (s *TechnicalAnalysisService) CalculateMACD(prices []float64) (float64, float64, float64) {
	if len(prices) < 26 {
		return 0, 0, 0
	}

	ema12 := s.CalculateEMA(prices, 12)
	ema26 := s.CalculateEMA(prices, 26)

	macd := ema12 - ema26

	// Tính signal line (EMA của MACD)
	var macdValues []float64
	for i := 26; i < len(prices); i++ {
		ema12 := s.CalculateEMA(prices[:i+1], 12)
		ema26 := s.CalculateEMA(prices[:i+1], 26)
		macdValues = append(macdValues, ema12-ema26)
	}

	signal := s.CalculateEMA(macdValues, 9)
	histogram := macd - signal

	return macd, signal, histogram
}

// analyzeVolume phân tích volume dựa trên ratio so với SMA
func (s *TechnicalAnalysisService) analyzeVolume(klines []models.KlineData) models.VolumeAnalysis {
	var volumes []float64
	for _, k := range klines {
		v, err := strconv.ParseFloat(k.Volume, 64)
		if err != nil {
			continue
		}
		volumes = append(volumes, v)
	}
	if len(volumes) < models.VOLUME_SMA_PERIOD+1 {
		return models.VolumeAnalysis{}
	}
	currentVolume := decimal.NewFromFloat(volumes[len(volumes)-1])
	var sum float64
	for i := len(volumes) - models.VOLUME_SMA_PERIOD - 1; i < len(volumes)-1; i++ {
		sum += volumes[i]
	}
	volumeSMA := sum / float64(models.VOLUME_SMA_PERIOD)
	volumeSMA21 := decimal.NewFromFloat(volumeSMA)
	var volumeSignal, volumeStrength, confirmation string
	var volumeRatio decimal.Decimal
	if volumeSMA > 0 {
		volumeRatio = currentVolume.Div(decimal.NewFromFloat(volumeSMA))
	} else {
		volumeRatio = decimal.Zero
	}
	if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_3X)) {
		volumeSignal = "🔥 VOLUME EXPLOSION"
		volumeStrength = "EXTREME"
		confirmation = "Tín hiệu Cực MẠNH - Breakout/Breakdown được xác nhận"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_2X)) {
		volumeSignal = "🚀 HIGH VOLUME SPIKE"
		volumeStrength = "STRONG"
		confirmation = "Tín hiệu MẠNH - Xu hướng được hỗ trợ tốt"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(models.VOLUME_SPIKE_1_5X)) {
		volumeSignal = "📈 ABOVE AVERAGE VOLUME"
		volumeStrength = "MODERATE"
		confirmation = "Tín hiệu TRUNG BÌNH - Có sự quan tâm tăng lên"
	} else if volumeRatio.GreaterThanOrEqual(decimal.NewFromFloat(1.0)) {
		volumeSignal = "🟡 NORMAL VOLUME"
		volumeStrength = "NORMAL"
		confirmation = "Volume bình thường - Theo dõi thêm"
	} else {
		volumeSignal = "📉 LOW VOLUME"
		volumeStrength = "WEAK"
		confirmation = "Volume thấp - Tín hiệu yếu, cẩn thận với fake move"
	}
	return models.VolumeAnalysis{
		CurrentVolume:  currentVolume,
		VolumeSMA21:    volumeSMA21,
		VolumeRatio:    volumeRatio,
		VolumeSignal:   volumeSignal,
		VolumeStrength: volumeStrength,
		Confirmation:   confirmation,
	}
}

// AnalyzeCrypto phân tích crypto với dữ liệu kline
func (s *TechnicalAnalysisService) AnalyzeCrypto(symbol string, klines []models.KlineData, interval string) (string, error) {
	if len(klines) == 0 {
		return "", fmt.Errorf("không có dữ liệu cho %s", symbol)
	}

	// Chuyển đổi giá đóng cửa sang float64
	var closePrices []float64
	for _, kline := range klines {
		price, err := strconv.ParseFloat(kline.Close, 64)
		if err != nil {
			continue
		}
		closePrices = append(closePrices, price)
	}

	if len(closePrices) < models.RSI_PERIOD+1 {
		return "", fmt.Errorf("không đủ dữ liệu để tính toán cho %s", symbol)
	}

	// Tính các chỉ báo
	currentPrice := closePrices[len(closePrices)-1]
	rsi := s.CalculateRSI(closePrices, models.RSI_PERIOD)
	ema9 := s.CalculateEMA(closePrices, models.EMA_SHORT)
	ema21 := s.CalculateEMA(closePrices, models.EMA_MEDIUM)
	ema50 := s.CalculateEMA(closePrices, models.EMA_LONG)

	// Phân tích volume
	volumeAnalysis := s.analyzeVolume(klines)

	analysis := models.TrendAnalysis{
		Signals: make([]string, 0),
	}

	// Phân tích RSI
	if rsi > 70 {
		analysis.Direction = "bearish"
	} else if rsi < 30 {
		analysis.Direction = "bullish"
	} else {
		analysis.Direction = "sideways"
	}

	// Hệ thống 3 EMA
	if currentPrice > ema9 && ema9 > ema21 && ema21 > ema50 {
		analysis.Direction = "bullish"
		// Điều chỉnh strength dựa trên volume
		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			analysis.Strength = "strong"
			analysis.Signals = append(analysis.Signals, "🚀 **STRONG BULLISH**: Giá > EMA9 > EMA21 > EMA50")
			analysis.Signals = append(analysis.Signals, "✅ Tất cả EMA đều hướng lên và xếp chồng đúng thứ tự")
			analysis.Signals = append(analysis.Signals, "🔥 Volume cao xác nhận xu hướng mạnh")
			analysis.Recommendation = "🟢 **MUA/GIỮ** - Xu hướng tăng mạnh được xác nhận bởi volume"
		} else {
			analysis.Strength = "moderate"
			analysis.Signals = append(analysis.Signals, "📈 **MODERATE BULLISH**: Giá > EMA9 > EMA21 > EMA50")
			analysis.Signals = append(analysis.Signals, "✅ Tất cả EMA đều hướng lên và xếp chồng đúng thứ tự")
			analysis.Signals = append(analysis.Signals, "⚠️ Volume thấp - Cần theo dõi thêm")
			analysis.Recommendation = "🟡 **CẨN THẬN MUA** - Xu hướng tăng nhưng volume chưa xác nhận"
		}

	} else if currentPrice < ema9 && ema9 < ema21 && ema21 < ema50 {
		analysis.Direction = "bearish"
		// Điều chỉnh strength dựa trên volume
		if volumeAnalysis.VolumeStrength == "EXTREME" || volumeAnalysis.VolumeStrength == "STRONG" {
			analysis.Strength = "strong"
			analysis.Signals = append(analysis.Signals, "⚠️ **STRONG BEARISH**: Giá < EMA9 < EMA21 < EMA50")
			analysis.Signals = append(analysis.Signals, "❌ Tất cả EMA đều hướng xuống và xếp chồng đúng thứ tự")
			analysis.Signals = append(analysis.Signals, "🔥 Volume cao xác nhận xu hướng giảm mạnh")
			analysis.Recommendation = "🔴 **BÁN/ĐỨNG NGOÀI** - Xu hướng giảm mạnh được xác nhận bởi volume"
		} else {
			analysis.Strength = "moderate"
			analysis.Signals = append(analysis.Signals, "📉 **MODERATE BEARISH**: Giá < EMA9 < EMA21 < EMA50")
			analysis.Signals = append(analysis.Signals, "❌ Tất cả EMA đều hướng xuống và xếp chồng đúng thứ tự")
			analysis.Signals = append(analysis.Signals, "⚠️ Volume thấp - Cần theo dõi thêm")
			analysis.Recommendation = "🟡 **CẨN THẬN BÁN** - Xu hướng giảm nhưng volume chưa xác nhận"
		}
	} else if currentPrice > ema9 && currentPrice > ema21 && ema21 > ema50 {
		analysis.Direction = "bullish"
		analysis.Strength = "moderate"
		analysis.Signals = append(analysis.Signals, "📈 **MODERATE BULLISH**: Giá trên EMA9 và EMA21")

		if ema9 > ema21 {
			analysis.Signals = append(analysis.Signals, "✅ EMA9 > EMA21 - Động lượng tăng tốt")
		} else {
			analysis.Signals = append(analysis.Signals, "⚠️ EMA9 < EMA21 - Động lượng chưa mạnh")
		}
		// Volume affects recommendation
		if volumeAnalysis.VolumeStrength == "STRONG" || volumeAnalysis.VolumeStrength == "EXTREME" {
			analysis.Recommendation = "🟢 **MUA** - Volume support xu hướng tăng"
		} else {
			analysis.Recommendation = "🟡 **CẨN THẬN MUA** - Thiếu volume chưa xác nhận"
		}
	} else if currentPrice < ema9 && currentPrice < ema21 && ema21 < ema50 {
		analysis.Direction = "bearish"
		analysis.Strength = "moderate"
		analysis.Signals = append(analysis.Signals, "📉 **MODERATE BEARISH**: Giá dưới EMA9 và EMA21")

		if ema9 < ema21 {
			analysis.Signals = append(analysis.Signals, "❌ EMA9 < EMA21 - Áp lực bán mạnh")
		} else {
			analysis.Signals = append(analysis.Signals, "⚠️ EMA9 > EMA21 - Áp lực bán đang giảm")
		}
		// Volume affects recommendation
		if volumeAnalysis.VolumeStrength == "STRONG" || volumeAnalysis.VolumeStrength == "EXTREME" {
			analysis.Recommendation = "🔴 **BÁN** - Volume support xu hướng giảm"
		} else {
			analysis.Recommendation = "🟡 **CẨN THẬN BÁN** - Thiếu volume chưa xác nhận"
		}
	} else {
		analysis.Direction = "sideways"
		analysis.Strength = "weak"
		analysis.Signals = append(analysis.Signals, "↔️ **SIDEWAYS**: EMA bị xoắn, giá dao động")

		// Check for potential breakout signals
		if currentPrice > ema50 {
			analysis.Signals = append(analysis.Signals, "🟢 Giá vẫn trên EMA50 - Xu hướng tăng dài hạn")
		} else if currentPrice < ema50 {
			analysis.Signals = append(analysis.Signals, "🔴 Giá dưới EMA50 - Xu hướng giảm dài hạn")
		}
		// Volume can signal upcoming breakout
		if volumeAnalysis.VolumeStrength == "STRONG" || volumeAnalysis.VolumeStrength == "EXTREME" {
			analysis.Recommendation = "⚡ **CHỜ BREAKOUT** - Volume cao có thể báo hiệu breakout sắp tới"
			analysis.Signals = append(analysis.Signals, "🔥 Volume tăng trong tích luỹ - Chuẩn bị breakout")
		} else {
			analysis.Recommendation = "⏳ **CHỜ TÍN HIỆU** - Thị trường đang tích luỹ"
		}
	}

	// Golden Cross: EMA9 crosses above EMA21 (in uptrend confirmed by EMA50)
	if ema9 > ema21 && ema21 > ema50 {
		analysis.Signals = append(analysis.Signals, "⭐ **GOLDEN CROSS**: EMA9 cắt lên EMA21 trong xu hướng tăng")

		// Death Cross: EMA9 crosses below EMA21 (in downtrend confirmed by EMA50)
	} else if ema9 < ema21 && ema21 < ema50 {
		analysis.Signals = append(analysis.Signals, "💀 **DEATH CROSS**: EMA9 cắt xuống EMA21 trong xu hướng giảm")
	}

	// EMA21/EMA50 major crossovers
	if ema21 > ema50 {
		analysis.Signals = append(analysis.Signals, "🔄 EMA21 > EMA50 - Xu hướng trung hạn tích cực")
	} else {
		analysis.Signals = append(analysis.Signals, "🔄 EMA21 < EMA50 - Xu hướng trung hạn tiêu cực")
	}

	// Nếu chưa có khuyến nghị rõ ràng
	if analysis.Recommendation == "" {
		if analysis.Direction == "bullish" {
			analysis.Recommendation = "🟢 **CÓ THỂ MUA/GIỮ** - Ưu tiên theo dõi thêm tín hiệu xác nhận"
		} else if analysis.Direction == "bearish" {
			analysis.Recommendation = "🔴 **CÂN NHẮC BÁN/ĐỨNG NGOÀI** - Ưu tiên chờ tín hiệu đảo chiều"
		} else {
			analysis.Recommendation = "🟡 **TRUNG LẬP** - Chờ tín hiệu rõ ràng hơn"
		}
	}

	// Tạo thông báo
	message := fmt.Sprintf("📊 **Phân tích kỹ thuật %s (%s)**\n\n", strings.ToUpper(symbol), strings.ToUpper(interval))
	message += fmt.Sprintf("💰 **Giá hiện tại:** $%s\n\n", utils.FormatPriceN(currentPrice, 4))

	// Block HỆ THỐNG 3 EMA
	message += fmt.Sprintf("📈 **EMA 9:** $%s\n", utils.FormatPriceN(ema9, 4))
	message += fmt.Sprintf("📊 **EMA 21:** $%s\n", utils.FormatPriceN(ema21, 4))
	message += fmt.Sprintf("📉 **EMA 50:** $%s\n", utils.FormatPriceN(ema50, 4))
	message += fmt.Sprintf("🎯 **Xu hướng:** %s (%s)\n\n", strings.ToUpper(analysis.Direction), strings.ToUpper(analysis.Strength))

	// Block tín hiệu 3 EMA
	message += "**✨ Tín hiệu:**\n"
	for _, signal := range analysis.Signals {
		message += fmt.Sprintf("- %s\n", signal)
	}

	// Block RSI
	message += fmt.Sprintf("\n**📈 RSI(%d):** %.2f\n", models.RSI_PERIOD, rsi)
	if rsi > 70 {
		message += "- 🔴 RSI: Quá mua (Overbought)\n"
	} else if rsi < 30 {
		message += "- 🟢 RSI: Quá bán (Oversold)\n"
	} else {
		message += "- 🟡 RSI: Trung tính\n"
	}

	// Block khuyến nghị tổng hợp
	message += "\n**💡 KHUYẾN NGHỊ TỔNG HỢP:**\n"
	message += fmt.Sprintf("- %s\n", analysis.Recommendation)
	// Add RSI confirmation/divergence
	if analysis.Direction == "bullish" && rsi < 30 {
		message += "• 🎯 **RSI oversold + xu hướng tăng** = Cơ hội mua tốt!\n"
	} else if analysis.Direction == "bearish" && rsi > 70 {
		message += "• ⚠️ **RSI overbought + xu hướng giảm** = Tín hiệu bán mạnh!\n"
	} else if analysis.Direction == "bullish" && rsi > 70 {
		message += "• 🟡 **Xu hướng tăng nhưng RSI cao** = Cẩn thận với pullback\n"
	} else if analysis.Direction == "bearish" && rsi < 30 {
		message += "• 🟡 **Xu hướng giảm nhưng RSI thấp** = Có thể bounce ngắn hạn\n"
	}

	// Block quản lý rủi ro
	message += "\n**⚠️ QUẢN LÝ RỦI RO:**\n"
	if analysis.Direction == "bullish" {
		message += fmt.Sprintf("• Stop-loss: Dưới EMA21 (~$%s)\n", utils.FormatPriceN(ema21, 4))
		message += "• Take-profit: Theo dõi RSI và EMA crossover\n"
	} else if analysis.Direction == "bearish" {
		message += fmt.Sprintf("• Stop-loss: Trên EMA21 (~$%s)\n", utils.FormatPriceN(ema21, 4))
		message += "• Target: Theo dõi support và EMA50\n"
	} else {
		message += "• Chờ breakout khỏi vùng tích luỹ\n"
		message += fmt.Sprintf("• Watch level: EMA50 ($%s)\n", utils.FormatPriceN(ema50, 4))
	}

	// Block Volume
	message += "\n**🔊 PHÂN TÍCH VOLUME:**\n"
	if !volumeAnalysis.CurrentVolume.IsZero() {
		message += fmt.Sprintf("- Volume hiện tại: %s\n", utils.FormatVolume(volumeAnalysis.CurrentVolume))
		message += fmt.Sprintf("- SMA %d Volume: %s\n", models.VOLUME_SMA_PERIOD, utils.FormatVolume(volumeAnalysis.VolumeSMA21))
		message += fmt.Sprintf("- Tỷ lệ Volume/SMA: x%s\n", volumeAnalysis.VolumeRatio.StringFixed(2))
		message += fmt.Sprintf("- %s (%s)\n", volumeAnalysis.VolumeSignal, volumeAnalysis.VolumeStrength)
		message += fmt.Sprintf("- %s\n", volumeAnalysis.Confirmation)
	} else {
		message += "- Không đủ dữ liệu volume để phân tích\n"
	}

	loc := time.FixedZone("UTC+7", 7*60*60)
	// Thời gian cập nhật
	message += fmt.Sprintf("\n⏰ **Cập nhật lúc:** %s", time.Now().In(loc).Format("15:04:05 02/01/2006"))

	return message, nil
}

// AnalyzeSignals phân tích tín hiệu từ các chỉ báo (giữ lại cho backward compatibility)
func (s *TechnicalAnalysisService) AnalyzeSignals(indicators models.TechnicalIndicators) []string {
	var signals []string

	// Phân tích RSI
	if indicators.RSI.GreaterThan(decimal.NewFromInt(70)) {
		signals = append(signals, "RSI: Quá mua (>70)")
	} else if indicators.RSI.LessThan(decimal.NewFromInt(30)) {
		signals = append(signals, "RSI: Quá bán (<30)")
	} else {
		signals = append(signals, "RSI: Trung tính (30-70)")
	}

	// Phân tích EMA
	if indicators.EMA20.GreaterThan(indicators.EMA50) && indicators.EMA50.GreaterThan(indicators.EMA200) {
		signals = append(signals, "EMA: Xu hướng tăng (Golden Cross)")
	} else if indicators.EMA20.LessThan(indicators.EMA50) && indicators.EMA50.LessThan(indicators.EMA200) {
		signals = append(signals, "EMA: Xu hướng giảm (Death Cross)")
	} else {
		signals = append(signals, "EMA: Xu hướng hỗn hợp")
	}

	// Phân tích MACD
	if indicators.MACD.GreaterThan(indicators.MACDSignal) {
		signals = append(signals, "MACD: Tín hiệu tăng")
	} else {
		signals = append(signals, "MACD: Tín hiệu giảm")
	}

	return signals
}

// AssessRisk đánh giá mức độ rủi ro (giữ lại cho backward compatibility)
func (s *TechnicalAnalysisService) AssessRisk(indicators models.TechnicalIndicators) string {
	riskScore := 0

	// RSI risk
	if indicators.RSI.GreaterThan(decimal.NewFromInt(80)) || indicators.RSI.LessThan(decimal.NewFromInt(20)) {
		riskScore += 3
	} else if indicators.RSI.GreaterThan(decimal.NewFromInt(70)) || indicators.RSI.LessThan(decimal.NewFromInt(30)) {
		riskScore += 2
	}

	// EMA risk
	if indicators.EMA20.LessThan(indicators.EMA50) && indicators.EMA50.LessThan(indicators.EMA200) {
		riskScore += 2
	}

	// MACD risk
	if indicators.MACD.LessThan(indicators.MACDSignal) {
		riskScore += 1
	}

	switch {
	case riskScore >= 5:
		return "CAO"
	case riskScore >= 3:
		return "TRUNG BÌNH"
	default:
		return "THẤP"
	}
}
