package database

import (
	"go-chain/config"
	"go-chain/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	if cfg.DSN != "" {
		// MySQL mode (cloud / production)
		dialector = mysql.Open(cfg.DSN)
	} else {
		// SQLite mode (local development)
		dialector = sqlite.Open(cfg.DBPath)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	// Auto-migrate tables
	err = DB.AutoMigrate(
		&models.Category{},
		&models.Recipe{},
		&models.Favorite{},
		&models.ShoppingItem{},
	)
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
