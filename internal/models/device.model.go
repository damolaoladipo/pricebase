package models

import (
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

type Device struct {
	UserID string `json:"userId" db:"user_id"`

	Login         string `json:"login" db:"login"`
	Platform      string `json:"platform" db:"platform"`
	DeviceVersion string `json:"deviceVersion" db:"device_version"`
	DeviceID      string `json:"deviceID" db:"device_id"`

	OS      utils.DeviceOS      `json:"os" db:"os"`
	Client  utils.DeviceClient  `json:"client" db:"client"`
	Details utils.DeviceDetails `json:"details" db:"details"`

	Source    string `json:"source" db:"source"`
	PushToken string `json:"pushToken" db:"push_token"`

	IsMobile  bool `json:"isMobile" db:"is_mobile"`
	IsDesktop bool `json:"isDesktop" db:"is_desktop"`

	Browser string `json:"browser" db:"browser"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
