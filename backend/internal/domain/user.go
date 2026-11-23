package domain

import "time"

type User struct {
	ID        uint `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	Password  *string // hashed, nil kalau daftar via Google saja
	GoogleID  *string `gorm:"uniqueIndex"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	FindByID(id uint) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByGoogleID(googleID string) (*User, error)
	Create(user *User) error
	Update(user *User) error
}
