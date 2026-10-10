package models

import (
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

type Settings struct {
	Code string `json:"code" db:"code"`

	Login   utils.LoginSettings   `json:"login" db:"login"`
	Usage   utils.UsageSettings   `json:"usage" db:"usage"`
	Billing utils.BillingSettings `json:"billing" db:"billing"`

	UserID     string `json:"userId,omitempty" db:"user_id"`
	BusinessID string `json:"businessId,omitempty" db:"business_id"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
