package domain

import "errors"

var (
	ErrEntryNotFound     = errors.New("entry tidak ditemukan")
	ErrWorkspaceNotFound = errors.New("workspace tidak ditemukan")
)
