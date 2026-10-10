package models

import (
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

type RoleTeam struct {
	Code        string   `json:"code" db:"code"`
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	Slug        string   `json:"slug" db:"slug"`
	Users       []string `json:"users" db:"users"`

	BusinessID  string                 `json:"businessId" db:"business_id"`
	Permissions []utils.UserPermission `json:"permissions" db:"permissions"`

	IsDefault bool `json:"isDefault" db:"is_default"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
