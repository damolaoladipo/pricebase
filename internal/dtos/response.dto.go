package dtos

type ResponseDTO[T any] struct {
	Error      bool           `json:"error"`
	Errors     []any          `json:"errors"`
	Count      *int64         `json:"count,omitempty"`
	Total      *int64         `json:"total,omitempty"`
	Pagination *PaginationDTO `json:"pagination,omitempty"`
	Data       T              `json:"data"`
	Message    string         `json:"message"`
	Status     int            `json:"status"`
	Token      *string        `json:"token,omitempty"`
}

type PaginationDTO struct {
	Next PaginationPageDTO `json:"next"`
	Prev PaginationPageDTO `json:"prev"`
}

type PaginationPageDTO struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
