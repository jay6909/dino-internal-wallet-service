package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID   uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name string    `gorm:"not null"`
	Role string    `gorm:"not null, default:'user'"` //system/user
	BaseTimeStamps
}

type UserRepository interface {
	GetUserByID(userID string) (*User, error)
	CreateUser(user *User) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) GetUserByID(userID string) (*User, error) {
	var user User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) CreateUser(user *User) error {
	return r.db.Create(user).Error
}
