package repository

type App struct {
	Id              int64  `xorm:"bigint 'id' autoincr pk notnull comment('主键')" json:"id"`
	AccessKeyId     string `xorm:"varchar(50) 'access_key_id' notnull default('') comment('访问秘钥ID')" json:"accessKeyId"`
	AccessKeySecret string `xorm:"varchar(50) 'access_key_secret' notnull default('') comment('访问秘钥')" json:"accessKeySecret"`
	Endpoint        string `xorm:"varchar(200) 'endpoint' notnull default('') comment('端点')" json:"endpoint"`
	InnerEndpoint   string `xorm:"varchar(200) 'inner_endpoint' notnull default('') comment('内部端点')" json:"innerEndpoint"`
	Status          int    `xorm:"int 'status' notnull default(1) comment('状态，0：禁用；1：正常')" json:"status"`
	DelStatus       int    `xorm:"int 'del_status' notnull default(0) comment('删除状态，0：未删除；1：已删除')" json:"-"`
	CreateTime      string `xorm:"datetime 'create_time' notnull comment('创建时间')" json:"createTime"`
	UpdateTime      string `xorm:"datetime 'update_time' notnull comment('更新时间')" json:"updateTime"`
}
