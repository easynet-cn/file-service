package object

import (
	"fmt"

	"github.com/easynet-cn/file-service/repository"
	"github.com/spf13/viper"
	"xorm.io/xorm"
)

var dbs map[string]*xorm.Engine

func InitDB(config *viper.Viper) {
	dbConfigs := config.GetStringMap("spring.datasources")
	dbs = make(map[string]*xorm.Engine, len(dbConfigs))

	for k := range dbConfigs {
		if engine, err := xorm.NewEngine("mysql", config.GetString(fmt.Sprintf("spring.datasources.%s.url", k))); err != nil {
			panic("连接数据库失败, error=" + err.Error())
		} else {
			engine.SetMaxOpenConns(config.GetInt(fmt.Sprintf("spring.datasources.%s.maxOpenConns", k)))
			engine.SetMaxIdleConns(config.GetInt(fmt.Sprintf("spring.datasources.%s.maxIdleConns", k)))

			dbs[k] = engine
		}
	}

}

func GetDatabases() map[string]*xorm.Engine {
	return dbs
}

func GetDB() *xorm.Engine {
	return dbs["file"]
}

func SyncDB() error {
	engine := GetDB()

	return engine.Sync2(
		&repository.App{},
		&repository.Bucket{},
		&repository.File{},
	)
}
