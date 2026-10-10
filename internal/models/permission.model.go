package models

import (
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

type Permission struct {
	ID        string                   `json:"id" db:"id"`
	Name      string                   `json:"name" db:"name"`
	Code      string                   `json:"code" db:"code"`
	Group     string                   `json:"group" db:"group"`
	Slug      string                   `json:"slug" db:"slug"`
	Actions   []utils.PermissionAction `json:"actions" db:"actions"`
	UpdatedBy string                   `json:"updatedBy,omitempty" db:"updated_by"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
}
