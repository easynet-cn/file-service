package object

import (
	"encoding/json"

	"github.com/easynet-cn/file-service/repository"
	"github.com/easynet-cn/file-service/util"
)

type Bucket struct {
	Id            int64          `json:"id"`
	AppId         int64          `json:"appId"`
	BucketType    int            `json:"bucketType"`
	Name          string         `json:"name"`
	Domain        string         `json:"domain"`
	ProcessConfig *ProcessConfig `json:"processConfig"`
	Status        int            `json:"status"`
	CreateTime    string         `json:"createTime"`
	UpdateTime    string         `json:"updateTime"`
}

func SearchBuckets(searchParam PageParam) (PageResult, error) {
	engine := GetDB()
	total := int64(0)

	if _, err := engine.SQL("SELECT COUNT(id) FROM bucket WHERE del_status=0").Get(&total); err != nil {
		return *NewPageResult(), err
	}

	if total > 0 {
		ms := make([]Bucket, 0)

		if err := engine.SQL("SELECT * FROM bucket WHERE del_status=0 LIMIT ?,?", searchParam.Start(), searchParam.PageSize).Find(&ms); err != nil {
			return *NewPageResult(), err
		}

		pageResult := &PageResult{Total: total, Data: make([]any, len(ms))}

		pageResult.TotalPages = pageResult.GetTotalPages(searchParam.PageSize)

		for i, m := range ms {
			pageResult.Data[i] = m
		}

		return *pageResult, nil
	} else {
		return PageResult{Data: make([]any, 0)}, nil
	}
}

func CreateBucket(m Bucket) (*Bucket, error) {
	entity := BucketToEntity(m)

	now := util.GetCurrentLocalDateTime()

	entity.CreateTime = now
	entity.UpdateTime = now

	if err := repository.BucketRepository.Create(GetDB(), entity); err != nil || entity.Id == 0 {
		return nil, err
	}

	return EntityToBucket(*entity), nil
}

func UpdateBucket(m Bucket) (*Bucket, error) {
	engine := GetDB()

	if bucketEntity, err := repository.BucketRepository.FindById(engine, m.Id); err != nil || bucketEntity.Id == 0 {
		return nil, err
	} else if cols := getUpdateBucketCols(bucketEntity, m); len(cols) == 0 {
		return EntityToBucket(*bucketEntity), nil
	} else {
		cols = append(cols, "update_time")

		bucketEntity.UpdateTime = util.GetCurrentLocalDateTime()

		if err := repository.BucketRepository.Update(engine, cols, bucketEntity); err != nil {
			return nil, err
		} else {
			return EntityToBucket(*bucketEntity), nil
		}
	}
}

func DeleteBucketById(id int64) (int64, error) {
	return repository.BucketRepository.DeleteById(GetDB(), id)
}

func BucketToEntity(m Bucket) *repository.Bucket {
	entity := &repository.Bucket{
		Id:         m.Id,
		AppId:      m.AppId,
		BucketType: m.BucketType,
		Name:       m.Name,
		Domain:     m.Domain,
		Status:     m.Status,
		CreateTime: m.CreateTime,
		UpdateTime: m.UpdateTime,
	}

	if m.ProcessConfig != nil {
		if bytes, err := json.Marshal(m.ProcessConfig); err != nil {
			entity.ProcessConfig = string(bytes)
		}
	}

	return entity
}

func EntityToBucket(entity repository.Bucket) *Bucket {
	m := &Bucket{
		Id:         entity.Id,
		AppId:      entity.AppId,
		BucketType: entity.BucketType,
		Name:       entity.Name,
		Domain:     entity.Domain,
		Status:     entity.Status,
		CreateTime: entity.CreateTime,
		UpdateTime: entity.UpdateTime,
	}

	if entity.ProcessConfig != "" {
		processConfig := &ProcessConfig{}

		if err := json.Unmarshal([]byte(entity.ProcessConfig), processConfig); err == nil {
			m.ProcessConfig = processConfig
		}
	}

	return m
}

func getUpdateBucketCols(entity *repository.Bucket, m Bucket) []string {
	cols := make([]string, 0)

	if entity.AppId != m.AppId {
		cols = append(cols, "app_id")

		entity.AppId = m.AppId
	}
	if entity.BucketType != m.BucketType {
		cols = append(cols, "bucket_type")

		entity.BucketType = m.BucketType
	}
	if entity.Name != m.Name {
		cols = append(cols, "name")

		entity.Name = m.Name
	}
	if entity.Domain != m.Domain {
		cols = append(cols, "domain")

		entity.Domain = m.Domain
	}
	if entity.Status != m.Status {
		cols = append(cols, "status")

		entity.Status = m.Status
	}

	return cols
}
