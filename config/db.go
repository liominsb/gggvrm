package config

import (
	"gggvrm/global"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() {
	dsn := Appconf.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()

	sqlDB.SetMaxIdleConns(Appconf.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(Appconf.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	global.Db = db
}
