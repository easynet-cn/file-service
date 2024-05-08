package object

type OssUploadToken struct {
	FileId      int64  `json:"fileId"`      //文件ID
	UploadUrl   string `json:"uploadUrl"`   //上传地址
	AccessKeyId string `json:"accessKeyId"` //OSS秘钥ID
	Policy      string `json:"policy"`      //策略
	Signature   string `json:"signature"`   //签名
	Key         string `json:"key"`         //文件键值
	Url         string `json:"url"`         //文件地址
}
