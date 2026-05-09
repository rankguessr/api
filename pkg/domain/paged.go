package domain

type Paged[T any] struct {
	Items      []T `json:"items"`
	PagesTotal int `json:"pages_total"`
}
