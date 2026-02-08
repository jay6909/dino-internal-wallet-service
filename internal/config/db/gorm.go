package config_db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB interface {
	GetDB() *gorm.DB
}
type dbGorm struct {
	DB *gorm.DB
}

func NewGormDB(dsn string) (*dbGorm, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return &dbGorm{DB: db}, err
}

func (d *dbGorm) GetDB() *gorm.DB {
	return d.DB
}
