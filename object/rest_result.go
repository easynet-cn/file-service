package object

type RestResult struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}
