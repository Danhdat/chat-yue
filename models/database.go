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

	// Cấu hình GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Kết nối database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("không thể kết nối database: %v", err)
	}

	// Cấu hình connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("không thể lấy connection pool: %v", err)
	}

	// Cấu hình connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
	)

	if err != nil {
		return fmt.Errorf("lỗi migrate database: %v", err)
	}

	log.Println("✅ Migrate database hoàn thành")
	return nil
}
