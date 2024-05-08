package repository

import (
	"github.com/easynet-cn/file-service/util"
	"xorm.io/xorm"
)

type bucketRepository struct{}

var BucketRepository = &bucketRepository{}

func (r *bucketRepository) FindById(engine *xorm.Engine, id int64) (*Bucket, error) {
	entity := &Bucket{}

	_, err := engine.ID(id).Where("del_status=0").Get(entity)

	return entity, err
}

func (r *bucketRepository) FindByName(engine *xorm.Engine, name string) (*Bucket, error) {
	entity := &Bucket{}

	_, err := engine.Where("name=? AND del_status=0", name).Get(entity)

	return entity, err
}

func (r *bucketRepository) FindByIdIn(engine *xorm.Engine, ids []int64) ([]Bucket, error) {
	entities := make([]Bucket, 0)

	err := engine.In("id", ids).Find(&entities)

	return entities, err
}

func (r *bucketRepository) Create(engine *xorm.Engine, entity *Bucket) error {
	_, err := engine.Insert(entity)

	return err
}

func (r *bucketRepository) Update(engine *xorm.Engine, cols []string, entity *Bucket) error {
	_, err := engine.ID(entity.Id).Cols(cols...).Update(entity)

	return err
}

func (r *bucketRepository) DeleteById(engine *xorm.Engine, id int64) (int64, error) {
	return engine.ID(id).Where("del_status=0").Update(&Bucket{DelStatus: 1, UpdateTime: util.GetCurrentLocalDateTime()})
}
