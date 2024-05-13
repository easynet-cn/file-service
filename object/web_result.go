package object

type WebResult struct {
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Data  any    `json:"data"`
	Total int64  `json:"total"`
}
