package domain

import "time"

// RefreshToken disimpan per device biar bisa di-revoke individual (logout per device).
type RefreshToken struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	TokenHash string `gorm:"uniqueIndex;not null"`
	DeviceInfo string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	FindByHash(hash string) (*RefreshToken, error)
	Create(rt *RefreshToken) error
	Revoke(id uint) error
	RevokeAllByUser(userID uint) error
}
