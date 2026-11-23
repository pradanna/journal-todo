package postgres

import (
	"time"

	"jurnal-be/internal/domain"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) FindByHash(hash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Create(rt *domain.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *refreshTokenRepository) Revoke(id uint) error {
	return r.db.Model(&domain.RefreshToken{}).Where("id = ?", id).Update("revoked_at", time.Now()).Error
}

func (r *refreshTokenRepository) RevokeAllByUser(userID uint) error {
	return r.db.Model(&domain.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error
}
