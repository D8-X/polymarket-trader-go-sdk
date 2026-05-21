package models

type PaginatedResponse[T any] struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor"`
	Count      int    `json:"count"`
	Data       []T    `json:"data"`
}
