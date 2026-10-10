package models

import (
	"time"
)

type Notification struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
