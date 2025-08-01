package models

import (
	"time"

	"github.com/google/uuid"
)

type RecordState int

const (
	PendingDelivery RecordState = iota
	Delivered
	MaxAttemptsReached
)

type Message struct {
	Key     string
	Headers map[string]string
	Body    []byte
	Topic   string
}

type Record struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Message          Message
	State            RecordState
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	LockID           *string
	LockedAt         *time.Time
	ProcessedAt      *time.Time
	NumberOfAttempts int
	LastAttemptAt    *time.Time
	Error            *string
}

type RawRecord struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Message          []byte
	State            RecordState
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	LockID           *string
	LockedAt         *time.Time
	ProcessedAt      *time.Time
	NumberOfAttempts int
	LastAttemptAt    *time.Time
	Error            *string
}
