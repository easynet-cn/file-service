package repository

import (
	"github.com/golang-module/carbon/v2"
	"xorm.io/xorm"
)

type appRepository struct{}

var AppRepository = &appRepository{}

func (r *appRepository) FindById(engine *xorm.Engine, id int64) (*App, error) {
	entity := &App{}

	_, err := engine.ID(id).Where("del_status=0").Get(entity)

	return entity, err
}

func (r *appRepository) FindByBucketName(engine *xorm.Engine, bucketName string) (*App, error) {
	entity := &App{}

	_, err := engine.SQL("SELECT a.* FROM app a JOIN bucket b ON a.id=b.app_id WHERE b.name=?", bucketName).Get(engine)

	return entity, err
}

func (r *appRepository) FindByIdIn(engine *xorm.Engine, ids []int64) ([]App, error) {
	entities := make([]App, 0)

	err := engine.In("id", ids).Find(&entities)

	return entities, err
}

func (r *appRepository) Create(engine *xorm.Engine, entity *App) error {
	_, err := engine.Insert(entity)

	return err
}

func (r *appRepository) Update(engine *xorm.Engine, cols []string, entity *App) error {
	_, err := engine.ID(entity.Id).Cols(cols...).Update(entity)

	return err
}

func (r *appRepository) DeleteById(engine *xorm.Engine, id int64) (int64, error) {
	return engine.ID(id).Where("del_status=0").Update(&App{DelStatus: 1, UpdateTime: carbon.Now().ToDateTimeString()})
}
