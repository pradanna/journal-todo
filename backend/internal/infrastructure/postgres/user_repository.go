package postgres

import (
	"jurnal-be/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByGoogleID(googleID string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("google_id = ?", googleID).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}
