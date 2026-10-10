package models

import "time"

type Role struct {
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	Slug        string   `json:"slug" db:"slug"`
	Permissions []string `json:"permissions" db:"permissions"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
