package object

type OssUploadFile struct {
	Bucket            string         `json:"bucket" form:"bucket"`                       //bucket名称
	Prefix            string         `json:"prefix" form:"prefix"`                       //文件前缀
	FileKey           string         `json:"fileKey" form:"fileKey"`                     //文件key
	SourceFile        string         `json:"sourceFile" form:"sourceFile"`               //源文件
	SourceFileSize    int64          `json:"sourceFileSize" form:"sourceFileSize"`       //源文件大小
	SourceFileType    string         `json:"sourceFileType" form:"sourceFileType"`       //源文件类型
	SourceFileAttr    string         `json:"sourceFileAttr" form:"sourceFileAttr"`       //源文件属性
	UseSourceFilename int            `json:"useSourceFilename" form:"useSourceFilename"` //是否使用源文件名
	ExpiredInSec      int64          `json:"expiredInSec" form:"expiredInSec"`           //过期秒数
	ProcessParams     []ProcessParam `json:"processParams"`                              //处理参数
	ProcessParamsStr  string         `form:"processParams"`                              //处理参数
}
