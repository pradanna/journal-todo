package application

import (
	"time"

	"jurnal-be/internal/domain"
)

type EntryService struct {
	entries    domain.JournalEntryRepository
	workspaces domain.WorkspaceRepository
}

func NewEntryService(entries domain.JournalEntryRepository, workspaces domain.WorkspaceRepository) *EntryService {
	return &EntryService{entries: entries, workspaces: workspaces}
}

type CreateEntryInput struct {
	UserID      uint
	WorkspaceID *uint
	EntryDate   time.Time
	HoursSpent  float64
	Mood        *string
	Items       []CreateEntryItemInput
}

type CreateEntryItemInput struct {
	Type     domain.EntryItemType
	Content  string
	Metadata string
}

func (s *EntryService) CreateEntry(input CreateEntryInput) (*domain.JournalEntry, error) {
	// kalau workspace diisi, pastikan memang milik user ini
	if input.WorkspaceID != nil {
		ws, err := s.workspaces.FindByID(*input.WorkspaceID)
		if err != nil || ws.UserID != input.UserID {
			return nil, domain.ErrWorkspaceNotFound
		}
	}

	items := make([]domain.EntryItem, 0, len(input.Items))
	for _, it := range input.Items {
		items = append(items, domain.EntryItem{
			Type:     it.Type,
			Content:  it.Content,
			Metadata: it.Metadata,
		})
	}

	entry := &domain.JournalEntry{
		UserID:      input.UserID,
		WorkspaceID: input.WorkspaceID,
		EntryDate:   input.EntryDate,
		HoursSpent:  input.HoursSpent,
		Mood:        input.Mood,
		Items:       items,
	}

	if err := s.entries.Create(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *EntryService) ListEntries(userID uint, filter domain.JournalEntryFilter) ([]domain.JournalEntry, error) {
	return s.entries.FindAllByUser(userID, filter)
}

func (s *EntryService) GetEntry(id uint, userID uint) (*domain.JournalEntry, error) {
	entry, err := s.entries.FindByID(id)
	if err != nil {
		return nil, err
	}
	if entry.UserID != userID {
		return nil, domain.ErrEntryNotFound
	}
	return entry, nil
}

func (s *EntryService) DeleteEntry(id uint, userID uint) error {
	entry, err := s.entries.FindByID(id)
	if err != nil {
		return err
	}
	if entry.UserID != userID {
		return domain.ErrEntryNotFound
	}
	return s.entries.Delete(id)
}
