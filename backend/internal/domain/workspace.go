package domain

import "time"

// Workspace adalah label bebas yang dibuat user sendiri (misal "Klien A", "Project X").
// Tidak ada nama/jumlah yang di-hardcode di sini.
type Workspace struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkspaceRepository interface {
	FindAllByUser(userID uint) ([]Workspace, error)
	FindByID(id uint) (*Workspace, error)
	Create(w *Workspace) error
	Update(w *Workspace) error
	Delete(id uint) error
}
