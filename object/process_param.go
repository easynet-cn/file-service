package object

type ProcessParam struct {
	Name   string   `json:"name" form:"name"`     //处理名称
	Params []string `json:"params" form:"params"` //处理参数
}
