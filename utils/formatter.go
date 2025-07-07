package utils

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// FormatVolume format volume với đơn vị K, M, B để dễ đọc
func FormatVolume(volume decimal.Decimal) string {
	// Chuyển sang float64 để tính toán
	value, _ := volume.Float64()

	if value == 0 {
		return "0"
	}

	// Xác định đơn vị và giá trị
	var unit string
	var displayValue float64

	switch {
	case value >= 1e9: // >= 1 tỷ
		unit = "B"
		displayValue = value / 1e9
	case value >= 1e6: // >= 1 triệu
		unit = "M"
		displayValue = value / 1e6
	case value >= 1e3: // >= 1 nghìn
		unit = "K"
		displayValue = value / 1e3
	default:
		// Dưới 1 nghìn, hiển thị nguyên giá trị
		return fmt.Sprintf("%.2f", value)
	}

	// Format với 2 chữ số thập phân
	return fmt.Sprintf("%.2f%s", displayValue, unit)
}

// FormatPrice format giá với dấu phẩy ngăn cách hàng nghìn
func FormatPrice(price decimal.Decimal) string {
	value, _ := price.Float64()
	if value < 1 && value > 0 {
		// Hiển thị 8 số thập phân cho giá nhỏ
		return fmt.Sprintf("%.8f", value)
	}
	// Format với dấu phẩy ngăn cách
	formatted := fmt.Sprintf("%.2f", value)
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	for i := len(integerPart) - 3; i > 0; i -= 3 {
		integerPart = integerPart[:i] + "," + integerPart[i:]
	}
	if len(parts) > 1 {
		return integerPart + "." + parts[1]
	}
	return integerPart
}

// FormatPercentage format phần trăm với dấu + hoặc -
func FormatPercentage(percent decimal.Decimal) string {
	value, _ := percent.Float64()

	if value > 0 {
		return fmt.Sprintf("+%.2f%%", value)
	} else if value < 0 {
		return fmt.Sprintf("%.2f%%", value)
	}
	return "0.00%"
}

// FormatPriceN format giá với n số thập phân và dấu phẩy ngăn cách
func FormatPriceN(value float64, n int) string {
	format := fmt.Sprintf("%%.%df", n)
	formatted := fmt.Sprintf(format, value)
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	for i := len(integerPart) - 3; i > 0; i -= 3 {
		integerPart = integerPart[:i] + "," + integerPart[i:]
	}
	if len(parts) > 1 {
		return integerPart + "." + parts[1]
	}
	return integerPart
}
