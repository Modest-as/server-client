package store

import "github.com/google/uuid"

// Store is common interface for all session stores
type Store interface {
	Remove(id uuid.UUID) error

	Exists(id uuid.UUID) bool

	Add(id uuid.UUID, n int) error

	GetSentCount(id uuid.UUID) int

	GetCurrentChecksum(id uuid.UUID) uint32

	HasNext(id uuid.UUID) (bool, error)

	GetNext(id uuid.UUID) uint32

	Update(id uuid.UUID, next uint32)
}
