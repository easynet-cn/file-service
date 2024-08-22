package object

import (
	"github.com/easynet-cn/file-service/repository"
	"github.com/easynet-cn/winter"
	"xorm.io/xorm"
)

var (
	Database *winter.Database
)

func GetDB() *xorm.Engine {
	return Database.GetDatabases()["file"]
}

func SyncDB() error {
	engine := GetDB()

	return engine.Sync2(
		&repository.App{},
		&repository.Bucket{},
		&repository.File{},
	)
}
