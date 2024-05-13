package object

import "math"

type PageResult struct {
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
	Data       []any `json:"data"`
}

func NewPageResult() *PageResult {
	return &PageResult{Total: 0, TotalPages: 0, Data: make([]any, 0)}
}

func (m *PageResult) GetTotalPages(pageSize int) int64 {
	return int64(math.Ceil(float64(m.Total) / float64(pageSize)))
}
