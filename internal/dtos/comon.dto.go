package dtos

import (
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// Infrastructure: what to bring up before the app can serve.
type ConnectAppDTO struct {
	Postgres *bool `json:"postgres,omitempty"`
	Migrate  *bool `json:"migrate,omitempty"`
	Seed     *bool `json:"seed,omitempty"`
	Redis    *bool `json:"redis,omitempty"`
	Memory   *bool `json:"memory,omitempty"`
}

// Runtime: what to start once infrastructure is ready.
type StartAppDTO struct {
	Server *bool `json:"server,omitempty"`
	Jobs   *bool `json:"jobs,omitempty"`
	Cron   *bool `json:"cron,omitempty"`
}

type LogRequestDTO = utils.LogRequestDTO

type LogDirectDTO = utils.LogDirectDTO

type RunScriptDTO struct {
	Type string `json:"type"`
	ID   any    `json:"id"`
}

type GetResourceDTO struct {
	Type string `json:"type"`
}
