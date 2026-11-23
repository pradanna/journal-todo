package postgres

import (
	"jurnal-be/internal/domain"
	"gorm.io/gorm"
)

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) FindAllByUser(userID uint) ([]domain.Workspace, error) {
	var ws []domain.Workspace
	err := r.db.Where("user_id = ?", userID).Order("name").Find(&ws).Error
	return ws, err
}

func (r *workspaceRepository) FindByID(id uint) (*domain.Workspace, error) {
	var w domain.Workspace
	if err := r.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *workspaceRepository) Create(w *domain.Workspace) error {
	return r.db.Create(w).Error
}

func (r *workspaceRepository) Update(w *domain.Workspace) error {
	return r.db.Save(w).Error
}

func (r *workspaceRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Workspace{}, id).Error
}
