package repository

type Bucket struct {
	Id            int64  `xorm:"bigint 'id' autoincr pk notnull comment('主键')" json:"id"`
	AppId         int64  `xorm:"bigint 'app_id' notnull default(0) comment('应用ID')" json:"appId"`
	BucketType    int    `xorm:"int 'bucket_type' notnull default(1) comment('空间类型，1：公有；2：私有')" json:"bucketType"`
	Name          string `xorm:"varchar(200) 'name' notnull default('') comment('名称')" json:"name"`
	Domain        string `xorm:"varchar(200) 'domain' notnull default('') comment('域名')" json:"domain"`
	ProcessConfig string `xorm:"text 'process_config' comment('处理配置')" json:"processConfig"`
	Status        int    `xorm:"int 'status' notnull default(1) comment('状态，0：禁用；1：正常')" json:"status"`
	DelStatus     int    `xorm:"int 'del_status' notnull default(0) comment('删除状态，0：未删除；1：已删除')" json:"-"`
	CreateTime    string `xorm:"datetime 'create_time' notnull comment('创建时间')" json:"createTime"`
	UpdateTime    string `xorm:"datetime 'update_time' notnull comment('更新时间')" json:"updateTime"`
}
