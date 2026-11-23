package postgres

import (
	"jurnal-be/internal/domain"
	"gorm.io/gorm"
)

type entryRepository struct {
	db *gorm.DB
}

func NewJournalEntryRepository(db *gorm.DB) domain.JournalEntryRepository {
	return &entryRepository{db: db}
}

func (r *entryRepository) FindByID(id uint) (*domain.JournalEntry, error) {
	var e domain.JournalEntry
	if err := r.db.Preload("Items").Preload("Tags").First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *entryRepository) FindAllByUser(userID uint, filter domain.JournalEntryFilter) ([]domain.JournalEntry, error) {
	q := r.db.Preload("Items").Preload("Tags").Where("user_id = ?", userID)
	if filter.From != nil {
		q = q.Where("entry_date >= ?", filter.From)
	}
	if filter.To != nil {
		q = q.Where("entry_date <= ?", filter.To)
	}
	if filter.WorkspaceID != nil {
		q = q.Where("workspace_id = ?", filter.WorkspaceID)
	}
	var entries []domain.JournalEntry
	err := q.Order("entry_date desc").Find(&entries).Error
	return entries, err
}

func (r *entryRepository) Create(entry *domain.JournalEntry) error {
	return r.db.Create(entry).Error
}

func (r *entryRepository) Update(entry *domain.JournalEntry) error {
	return r.db.Save(entry).Error
}

func (r *entryRepository) Delete(id uint) error {
	return r.db.Delete(&domain.JournalEntry{}, id).Error
}
