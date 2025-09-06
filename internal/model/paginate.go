package model

type Paginate[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Pages int `json:"pages"`
}
