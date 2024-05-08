package repository

import "xorm.io/xorm"

type fileRepository struct{}

var FileRepository = &fileRepository{}

func (r *fileRepository) Create(engine *xorm.Engine, entity *File) error {
	_, err := engine.Insert(entity)

	return err
}
