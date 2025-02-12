package object

import (
	"github.com/dromara/carbon/v2"
	"github.com/easynet-cn/file-service/repository"
	"github.com/easynet-cn/winter"
	"github.com/easynet-cn/winter/orm"
)

type App struct {
	Id              int64  `json:"id"`
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Endpoint        string `json:"endpoint"`
	InnerEndpoint   string `json:"innerEndpoint"`
	Status          int    `json:"status"`
	CreateTime      string `json:"createTime"`
	UpdateTime      string `json:"updateTime"`
}

func SearchApps(searchParam winter.PageParam) (winter.PageResult, error) {
	engine := GetDB()
	ms := make([]App, 0)

	if total, err := orm.FindPagination(engine, &ms, "SELECT COUNT(id) FROM app WHERE del_status=0", []any{}, "SELECT * FROM app WHERE del_status=0 LIMIT ?,?", []any{searchParam.Start(), searchParam.PageSize}); err != nil {
		return *winter.NewPageResult(), err
	} else {
		pageResult := &winter.PageResult{Total: total, Data: make([]any, len(ms))}

		pageResult.TotalPages = pageResult.GetTotalPages(searchParam.PageSize)

		for i, m := range ms {
			pageResult.Data[i] = m
		}

		return *pageResult, nil
	}
}

func CreateApp(m App) (*App, error) {
	entity := AppToEntity(m)

	now := carbon.Now().ToDateTimeString()

	entity.CreateTime = now
	entity.UpdateTime = now

	if err := orm.Create(GetDB(), entity); err != nil || entity.Id == 0 {
		return nil, err
	}

	return EntityToApp(*entity), nil
}

func UpdateApp(m App) (*App, error) {
	engine := GetDB()

	if appEntity, err := repository.AppRepository.FindById(engine, m.Id); err != nil || appEntity.Id == 0 {
		return nil, err
	} else if cols := getUpdateAppCols(appEntity, m); len(cols) == 0 {
		return EntityToApp(*appEntity), nil
	} else {
		cols = append(cols, "update_time")

		appEntity.UpdateTime = carbon.Now().ToDateTimeString()

		if err := repository.AppRepository.Update(engine, cols, appEntity); err != nil {
			return nil, err
		} else {
			return EntityToApp(*appEntity), nil
		}
	}
}

func DeleteAppById(id int64) (int64, error) {
	return repository.AppRepository.DeleteById(GetDB(), id)
}

func AppToEntity(m App) *repository.App {
	return &repository.App{
		Id:              m.Id,
		AccessKeyId:     m.AccessKeyId,
		AccessKeySecret: m.AccessKeySecret,
		Endpoint:        m.Endpoint,
		InnerEndpoint:   m.InnerEndpoint,
		Status:          m.Status,
		CreateTime:      m.CreateTime,
		UpdateTime:      m.UpdateTime,
	}
}

func EntityToApp(entity repository.App) *App {
	return &App{
		Id:              entity.Id,
		AccessKeyId:     entity.AccessKeyId,
		AccessKeySecret: entity.AccessKeySecret,
		Endpoint:        entity.Endpoint,
		InnerEndpoint:   entity.InnerEndpoint,
		Status:          entity.Status,
		CreateTime:      entity.CreateTime,
		UpdateTime:      entity.UpdateTime,
	}
}

func getUpdateAppCols(entity *repository.App, m App) []string {
	cols := make([]string, 0)

	if entity.AccessKeyId != m.AccessKeyId {
		cols = append(cols, "access_key_id")

		entity.AccessKeyId = m.AccessKeyId
	}
	if entity.AccessKeySecret != m.AccessKeySecret {
		cols = append(cols, "access_key_secret")

		entity.AccessKeySecret = m.AccessKeySecret
	}
	if entity.Endpoint != m.Endpoint {
		cols = append(cols, "endpoint")

		entity.Endpoint = m.Endpoint
	}
	if entity.InnerEndpoint != m.InnerEndpoint {
		cols = append(cols, "inner_endpoint")

		entity.InnerEndpoint = m.InnerEndpoint
	}
	if entity.Status != m.Status {
		cols = append(cols, "status")

		entity.Status = m.Status
	}

	return cols
}
