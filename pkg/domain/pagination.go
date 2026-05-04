package domain

type PagedResult[T any] struct {
	Items      []T `json:"items"`
	PagesTotal int `json:"pages_total"`
}
