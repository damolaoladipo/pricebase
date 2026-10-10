package utils

type Result[T any] struct {
	Error      bool        `json:"error"`
	Errors     []T         `json:"errors,omitempty"`
	Count      int         `json:"count,omitempty"`
	Total      int64       `json:"total,omitempty"`
	Message    string      `json:"message"`
	Code       int         `json:"code"`
	Data       T           `json:"data"`
	Report     *APIReport  `json:"report,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Token      string      `json:"token,omitempty"`
	Status     int         `json:"status,omitempty"`
	Filters    any         `json:"filters,omitempty"`
}
