package models

import (
	"fmt"
	"log"
	"time"

	"chatbtc/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase khởi tạo kết nối database
func InitDatabase() error {
	cfg := config.AppConfig

	// Tạo connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable ",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Cấu hình GORM logger - Tối ưu hóa cho production
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Chỉ log warnings và errors
		// Tối ưu hóa performance
		SkipDefaultTransaction: true, // Bỏ qua transaction mặc định cho các operation đơn giản
		PrepareStmt:            true, // Cache prepared statements
	}

	// Kết nối database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("không thể kết nối database: %v", err)
	}

	// Cấu hình connection pool - Tối ưu hóa cực đoan cho high performance
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("không thể lấy connection pool: %v", err)
	}

	// Cấu hình connection pool tối ưu cực đoan
	sqlDB.SetMaxIdleConns(50)                  // Tăng cao số connection idle
	sqlDB.SetMaxOpenConns(500)                 // Tăng cao số connection tối đa
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Giảm thời gian connection
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Giảm thời gian idle

	log.Println("✅ Kết nối database PostgreSQL thành công")
	return nil
}

// CloseDatabase đóng kết nối database
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate tự động tạo/migrate các bảng
func AutoMigrate() error {
	log.Println("🔄 Đang migrate database...")

	err := DB.AutoMigrate(
		&AnalysisRecord{},
		&PriceHistory{},
		&Symbol{},
		&DataUpdate{},
		&AutoVolumeRecord{},
		&NotificationLog{},
		&AlphaSymbol{},
		&HolderHistory{},
	)

	if err != nil {
		return fmt.Errorf("lỗi migrate database: %v", err)
	}

	log.Println("✅ Migrate database hoàn thành")
	return nil
}
