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

	// Tạo connection string với các tham số tối ưu cho Railway
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require "+
		"connect_timeout=10 "+
		"statement_timeout=30000 "+
		"idle_in_transaction_session_timeout=30000 "+
		"application_name=cryptobot",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	// Cấu hình GORM logger - Tối ưu hóa cho Railway production
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // Chỉ log errors để giảm overhead
		// Tối ưu hóa performance cho Railway
		SkipDefaultTransaction: true, // Bỏ qua transaction mặc định cho các operation đơn giản
		PrepareStmt:            true, // Cache prepared statements
		// Thêm timeout cho các operation
		NowFunc: func() time.Time {
			return time.Now().UTC() // Sử dụng UTC để tránh timezone issues
		},
	}

	// Kết nối database với retry logic
	var err error
	var retries = 3
	for i := 0; i < retries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			break
		}
		log.Printf("Lần thử kết nối %d thất bại: %v", i+1, err)
		if i < retries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("không thể kết nối database sau %d lần thử: %v", retries, err)
	}

	// Cấu hình connection pool - Tối ưu hóa cực đoan cho Railway
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("không thể lấy connection pool: %v", err)
	}

	// Cấu hình connection pool tối ưu cho Railway
	sqlDB.SetMaxIdleConns(5)                  // Giảm idle connections cho Railway
	sqlDB.SetMaxOpenConns(25)                 // Giảm max connections cho Railway
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Giảm thời gian connection
	sqlDB.SetConnMaxIdleTime(1 * time.Minute) // Giảm thời gian idle

	log.Println("✅ Kết nối database PostgreSQL thành công (Railway optimized)")
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
