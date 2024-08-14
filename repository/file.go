package repository

type File struct {
	Id             int64  `xorm:"bigint 'id' autoincr pk notnull comment('ID')" json:"id"`
	BucketId       int64  `xorm:"bigint 'bucket_id' notnull default(0) comment('空间ID')" json:"bucketId"`
	FileKey        string `xorm:"varchar(500) 'file_key' notnull default('') index comment('文件键值')" json:"fileKey"`
	SourceFile     string `xorm:"varchar(1000) 'source_file' notnull default('') comment('原文件')" json:"sourceFile"`
	SourceFileSize int64  `xorm:"bigint 'source_file_size' notnull default(0) comment('原文件大小')" json:"sourceFileSize"`
	SourceFileType string `xorm:"varchar(50) 'source_file_type' notnull default('') comment('原文件类型')" json:"sourceFileType"`
	SourceFileAttr string `xorm:"varchar(3000) 'source_file_attr' notnull default('') comment('原文件属性')" json:"sourceFileAttr"`
	DelStatus      int    `xorm:"int 'del_status' notnull default(0) comment('删除状态，0：未删除；1：已删除')" json:"-"`
	CreateTime     string `xorm:"datetime 'create_time' notnull comment('创建时间')" json:"createTime"`
	UpdateTime     string `xorm:"datetime 'update_time' notnull comment('更新时间')" json:"updateTime"`
}

func (*File) TableComment() string {
	return "文件"
}
