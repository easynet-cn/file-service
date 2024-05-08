package object

type PageParam struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

func (pageParam *PageParam) Start() int {
	return (pageParam.PageIndex - 1) * pageParam.PageSize
}
