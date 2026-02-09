package config_db

import (
	"github.com/jay6909/dino-internal-wallet-service/internal/data/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB interface {
	GetDB() *gorm.DB
	Migrate()
}

func (d *dbGorm) Migrate() {
	if d.db == nil {
		panic("gorm db is nil — migration aborted")
	}

	if err := d.db.AutoMigrate(
		&repository.User{},
		&repository.Wallet{},
		&repository.WalletTransaction{},
		&repository.CurrencyType{},
	); err != nil {
		panic(err)
	}
}

type dbGorm struct {
	db *gorm.DB
}

func NewGormDB(dsn string) (DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return &dbGorm{db: db}, err
}

func (d *dbGorm) GetDB() *gorm.DB {
	return d.db
}
