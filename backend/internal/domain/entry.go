package domain

import "time"

type EntryItemType string

const (
	EntryItemTask     EntryItemType = "task"
	EntryItemBlocker  EntryItemType = "blocker"
	EntryItemDecision EntryItemType = "decision"
	EntryItemTodo     EntryItemType = "todo"
)

// EntryItem polymorphic by Type, Metadata nampung field tambahan tanpa migration baru.
type EntryItem struct {
	ID             uint `gorm:"primaryKey"`
	JournalEntryID uint `gorm:"index;not null"`
	Type           EntryItemType `gorm:"type:varchar(20);not null"`
	Content        string `gorm:"type:text;not null"`
	Metadata       string `gorm:"type:jsonb"` // raw JSON, di-parse di application layer
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Tag struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"index;not null"`
	Name   string `gorm:"not null"`
}

type JournalEntry struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint `gorm:"index;not null"`
	WorkspaceID *uint `gorm:"index"`
	EntryDate   time.Time `gorm:"type:date;not null"`
	HoursSpent  float64
	Mood        *string
	Items       []EntryItem `gorm:"foreignKey:JournalEntryID"`
	Tags        []Tag `gorm:"many2many:entry_tags;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type JournalEntryFilter struct {
	From        *time.Time
	To          *time.Time
	WorkspaceID *uint
	Tag         *string
}

type JournalEntryRepository interface {
	FindByID(id uint) (*JournalEntry, error)
	FindAllByUser(userID uint, filter JournalEntryFilter) ([]JournalEntry, error)
	Create(entry *JournalEntry) error
	Update(entry *JournalEntry) error
	Delete(id uint) error
}
