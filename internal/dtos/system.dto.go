package dtos

import (
	types "github.com/damolaoladipo/common-go/types"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

type LogAuditRequestDTO = utils.LogAuditRequestDTO

type LogAPICallDTO = utils.LogAPICallDTO

type LogAuditDTO = utils.LogAuditDTO

type UpdateNotificationsDTO struct {
	SMS       bool `json:"sms"`
	Email     bool `json:"email"`
	Push      bool `json:"push"`
	Dashboard bool `json:"dashboard"`
}

type TestSMSDTO struct {
	PhoneNumber string `json:"phoneNumber"`
	Driver      string `json:"driver"`
	Message     string `json:"message"`
}

type EncryptDataDTO struct {
	Payload   any    `json:"payload"`
	Password  string `json:"password"`
	Separator string `json:"separator"`
}

type DecryptDataDTO struct {
	Payload   any    `json:"payload"`
	Password  string `json:"password"`
	Separator string `json:"separator"`
}

type FormatDateRangeDTO struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type CreateNotifyDTO struct {
	UserID      any    `json:"userId,omitempty"`
	BusinessID  any    `json:"businessId,omitempty"`
	Title       string `json:"title,omitempty"`
	Content     string `json:"content,omitempty"`
	Category    string `json:"category,omitempty"`
	Channel     string `json:"channel,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Destination string `json:"destination,omitempty"`
	Source      string `json:"source,omitempty"`
	MessageID   string `json:"messageID,omitempty"`
	PushToken   string `json:"pushToken,omitempty"`
	Metadata    any    `json:"metadata,omitempty"`
	URL         string `json:"url,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	ID          any    `json:"id,omitempty"`
}

type SyncNotifyMQDTO struct {
	Type   types.SyncType   `json:"type"`
	Notify types.SyncNotify `json:"notify"`
	Config CreateNotifyDTO  `json:"config"`
}

type SyncEventDTO struct {
	Type   types.SyncType   `json:"type"`
	Action types.SyncAction `json:"action"`
	Data   any              `json:"data"`
}
