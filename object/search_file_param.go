package object

import "github.com/easynet-cn/winter"

type SearchFileParam struct {
	Ids           []int64        `json:"ids"`           //文件ID集合
	FileKeys      []string       `json:"fileKeys"`      //文件key集合
	Buckets       []string       `json:"buckets"`       //bucket集合
	ExpiredInSec  int64          `json:"expiredInSec"`  //过期秒数
	ProcessParams []ProcessParam `json:"processParams"` //处理参数
}

type SearchFilePageParam struct {
	winter.PageParam
	SearchFileParam
}
