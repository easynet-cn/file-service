package object

type ProcessConfig struct {
	Expression    string         `json:"expression"`    // 表达式
	ProcessParams []ProcessParam `json:"processParams"` //处理参数
}
