package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"unique"`
	Email        string    `gorm:"unique"`
	Password     string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	RefreshToken string    `gorm:"not null;default:''"`
}
