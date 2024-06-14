package object

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/repository"
	"github.com/easynet-cn/file-service/util"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

type File struct {
	Id             int64  `json:"id"`
	BucketId       int64  `json:"bucketId"`
	BucketName     string `json:"bucketName"`
	Domain         string `json:"domain"`
	FileKey        string `json:"fileKey"`
	SourceFile     string `json:"sourceFile"`
	SourceFileSize int64  `json:"sourceFileSize"`
	SourceFileType string `json:"sourceFileType"`
	SourceFileAttr string `json:"sourceFileAttr"`
	Url            string `json:"url"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

var (
	ossClientCache = &sync.Map{}
	functions      = map[string]govaluate.ExpressionFunction{
		"hasPrefix": func(args ...any) (any, error) {
			return strings.HasPrefix(args[0].(string), args[1].(string)), nil
		},
		"hasSuffix": func(args ...any) (any, error) {
			return strings.HasSuffix(args[0].(string), args[1].(string)), nil
		},
		"toLower": func(args ...any) (any, error) {
			return strings.ToLower(args[0].(string)), nil
		},
		"toUpper": func(args ...any) (any, error) {
			return strings.ToUpper(args[0].(string)), nil
		},
	}
)

func UploadFile(uploadFile OssUploadFile, file string) (*File, error) {
	defer func(file string) {
		os.Remove(file)
	}(file)

	engine := GetDB()

	if ossBucket, err := repository.BucketRepository.FindByName(engine, uploadFile.Bucket); err != nil || ossBucket.Id == 0 {
		log.Logger.Error("repository.FindBucketByName", zap.String("bucketName", uploadFile.Bucket), zap.Error(err))

		return nil, err
	} else if appEntity, err := repository.AppRepository.FindById(engine, ossBucket.AppId); err != nil || appEntity.Id == 0 {
		log.Logger.Error("repository.FindAppById", zap.Any("appId", ossBucket.AppId), zap.Error(err))

		return nil, err
	} else if ossClient, err := getOssClientByBucket(*appEntity); err != nil {
		log.Logger.Error("getOssClientByBucket", zap.Any("appEntity", appEntity), zap.Error(err))

		return nil, err
	} else {
		fileKey := generateFileKey(uploadFile)

		fileEntity := &repository.File{
			BucketId:       ossBucket.Id,
			FileKey:        fileKey,
			SourceFile:     uploadFile.SourceFile,
			SourceFileType: uploadFile.SourceFileType,
			SourceFileSize: uploadFile.SourceFileSize,
			SourceFileAttr: uploadFile.SourceFileAttr,
		}

		fileEntity.BucketId = ossBucket.Id
		fileEntity.FileKey = fileKey
		fileEntity.SourceFile = uploadFile.SourceFile
		fileEntity.SourceFileSize = uploadFile.SourceFileSize

		now := util.GetCurrentLocalDateTime()

		fileEntity.CreateTime = now
		fileEntity.UpdateTime = now

		if err := repository.FileRepository.Create(engine, fileEntity); err != nil || fileEntity.Id == 0 {
			log.Logger.Error("repository.CreateFile", zap.Any("fileEntity", fileEntity), zap.Error(err))

			return nil, err
		}

		if bucket, err := ossClient.Bucket(ossBucket.Name); err != nil {
			log.Logger.Error("ossClient.Bucket", zap.String("bucketName", ossBucket.Name), zap.Error(err))

			return nil, err
		} else if err := ossUploadFile(bucket, fileKey, file); err != nil {
			log.Logger.Error("bucket.UploadFile", zap.String("fileKey", fileKey), zap.String("file", file), zap.Error(err))

			return nil, err
		} else {
			return &File{
				Id:             fileEntity.Id,
				BucketId:       ossBucket.Id,
				BucketName:     ossBucket.Name,
				Domain:         ossBucket.Domain,
				FileKey:        fileKey,
				SourceFile:     uploadFile.SourceFile,
				SourceFileSize: uploadFile.SourceFileSize,
				SourceFileType: uploadFile.SourceFileType,
				SourceFileAttr: uploadFile.SourceFileAttr,
				Url:            getUrl(ossClient, *appEntity, *ossBucket, fileKey, uploadFile.ExpiredInSec, uploadFile.ProcessParams),
			}, nil
		}
	}
}

func GetUploadToken(uploadFile OssUploadFile) (*OssUploadToken, error) {
	engine := GetDB()

	if ossBucket, err := repository.BucketRepository.FindByName(engine, uploadFile.Bucket); err != nil || ossBucket.Id == 0 {
		return nil, err
	} else if appEntity, err := repository.AppRepository.FindById(engine, ossBucket.AppId); err != nil || appEntity.Id == 0 {
		return nil, err
	} else if ossClient, err := getOssClientByBucket(*appEntity); err != nil {
		return nil, err
	} else {
		fileKey := generateFileKey(uploadFile)

		fileEntity := &repository.File{
			BucketId:       ossBucket.Id,
			FileKey:        fileKey,
			SourceFile:     uploadFile.SourceFile,
			SourceFileType: uploadFile.SourceFileType,
			SourceFileSize: uploadFile.SourceFileSize,
			SourceFileAttr: uploadFile.SourceFileAttr,
		}

		fileEntity.BucketId = ossBucket.Id
		fileEntity.FileKey = fileKey
		fileEntity.SourceFile = uploadFile.SourceFile
		fileEntity.SourceFileSize = uploadFile.SourceFileSize

		now := util.GetCurrentLocalDateTime()

		fileEntity.CreateTime = now
		fileEntity.UpdateTime = now

		if err := repository.FileRepository.Create(engine, fileEntity); err != nil || fileEntity.Id == 0 {
			return nil, err
		}

		bucket := ossBucket.Name
		endpoint := appEntity.Endpoint
		accessKeyId := appEntity.AccessKeyId
		secretAccessKey := appEntity.AccessKeySecret
		uploadUrl := fmt.Sprintf("//%s.%s", bucket, endpoint)

		if ossBucket.Domain != "" {
			uploadUrl = fmt.Sprintf("//%s", ossBucket.Domain)
		}

		expiredInSec := int64(60 * 60)
		expiration := time.Now().Add(time.Duration(60 * time.Minute)).UTC().Format(time.RFC3339Nano)

		if uploadFile.ExpiredInSec > 0 {
			expiredInSec = uploadFile.ExpiredInSec
			expiration = time.Now().Add(time.Duration(uploadFile.ExpiredInSec) * time.Second).UTC().Format(time.RFC3339Nano)
		}

		policyStr := fmt.Sprintf(`{"expiration":"%s","conditions":[{"bucket":"%s"},["eq","$key","%s"]]}`, expiration, bucket, fileKey)

		policy := base64.StdEncoding.EncodeToString([]byte(policyStr))

		key := []byte(secretAccessKey)
		mac := hmac.New(sha1.New, key)
		mac.Write([]byte(policy))

		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		return &OssUploadToken{
			FileId:      fileEntity.Id,
			UploadUrl:   uploadUrl,
			AccessKeyId: accessKeyId,
			Policy:      policy,
			Signature:   signature,
			Key:         fileKey,
			Url:         getUrl(ossClient, *appEntity, *ossBucket, fileKey, expiredInSec, uploadFile.ProcessParams)}, nil
	}
}

func SearchFiles(searchParam SearchFileParam) ([]File, error) {
	ms := make([]File, 0)

	if len(searchParam.Ids) == 0 && len(searchParam.FileKeys) == 0 {
		return ms, nil
	}

	engine := GetDB()
	sb := new(strings.Builder)
	params := make([]any, 0, len(searchParam.Ids)+len(searchParam.FileKeys))

	sb.WriteString("SELECT f.*,b.name AS bucket_name,b.domain FROM file f JOIN bucket b ON f.bucket_id=b.id JOIN app a ON b.app_id=a.id")
	sb.WriteString(" WHERE f.del_status=0 AND b.del_status=0 AND a.del_status=0")

	if len(searchParam.Ids) > 0 {
		sb.WriteString(" AND f.id IN(")

		for i, id := range searchParam.Ids {
			sb.WriteString("?")

			if i < len(searchParam.Ids)-1 {
				sb.WriteString(",")
			}

			params = append(params, id)
		}

		sb.WriteString(")")
	}

	if len(searchParam.FileKeys) > 0 {
		sb.WriteString(" AND f.file_key IN(")

		for i, fileKey := range searchParam.FileKeys {
			sb.WriteString("?")

			if i < len(searchParam.FileKeys)-1 {
				sb.WriteString(",")
			}

			params = append(params, fileKey)
		}

		sb.WriteString(")")
	}

	if len(searchParam.Buckets) > 0 {
		sb.WriteString(" AND b.name IN(")

		for i, bucket := range searchParam.Buckets {
			sb.WriteString("?")

			if i < len(searchParam.Buckets)-1 {
				sb.WriteString(",")
			}

			params = append(params, bucket)
		}

		sb.WriteString(")")
	}

	if err := engine.SQL(sb.String(), params...).Find(&ms); err != nil {
		return nil, err
	}

	expiredInSec := int64(60 * 60)

	if searchParam.ExpiredInSec > 0 {
		expiredInSec = searchParam.ExpiredInSec
	}

	if err := mergeFiles(engine, &ms, expiredInSec, searchParam.ProcessParams); err != nil {
		return nil, err
	}

	return ms, nil
}

func SearchPageFiles(searchParam SearchFilePageParam) (PageResult, error) {
	engine := GetDB()
	where, params := buildSearchFilesWhere(searchParam)
	countSb := new(strings.Builder)

	countSb.WriteString("SELECT COUNT(f.id) FROM file f JOIN bucket b ON f.bucket_id=b.id JOIN app a ON b.app_id=a.id")
	countSb.WriteString(where)

	total := int64(0)

	if _, err := engine.SQL(countSb.String(), params...).Get(&total); err != nil {
		return *NewPageResult(), err
	}

	if total > 0 {
		querySb := new(strings.Builder)
		queryParams := append(params, searchParam.Start(), searchParam.PageSize)

		querySb.WriteString("SELECT f.*,b.name AS bucket_name,b.domain FROM file f JOIN bucket b ON f.bucket_id=b.id JOIN app a ON b.app_id=a.id")
		querySb.WriteString(where)
		querySb.WriteString(" LIMIT ?,?")

		ms := make([]File, 0)

		if err := engine.SQL(querySb.String(), queryParams...).Find(&ms); err != nil {
			return *NewPageResult(), err
		}

		expiredInSec := int64(60 * 60)

		if searchParam.ExpiredInSec > 0 {
			expiredInSec = searchParam.ExpiredInSec
		}

		if err := mergeFiles(engine, &ms, expiredInSec, searchParam.ProcessParams); err != nil {
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

func getOssClientByBucket(appEntity repository.App) (*oss.Client, error) {
	if v, ok := ossClientCache.Load(strconv.FormatInt(appEntity.Id, 10)); ok {
		return v.(*oss.Client), nil
	} else if ossClient, err := oss.New(appEntity.Endpoint, appEntity.AccessKeyId, appEntity.AccessKeySecret); err != nil {
		return nil, err
	} else {
		ossClientCache.Store(strconv.FormatInt(appEntity.Id, 10), ossClient)

		return ossClient, nil
	}
}

func generateFileKey(uploadFile OssUploadFile) string {
	sb := new(strings.Builder)

	if uploadFile.FileKey != "" {
		sb.WriteString(strings.ToLower(uploadFile.FileKey))
	} else {
		if uploadFile.Prefix != "" {
			sb.WriteString(strings.ToLower(uploadFile.Prefix))
			sb.WriteString("/")
		}

		if uploadFile.UseSourceFilename == 1 {
			sb.WriteString(uploadFile.SourceFile)
		} else {
			sb.WriteString(NewObjectID().Hex())
			sb.WriteString(strings.ToLower(filepath.Ext(uploadFile.SourceFile)))
		}

	}

	return sb.String()
}

func getUrl(ossClient *oss.Client, app repository.App, bucket repository.Bucket, fileKey string, expiredInSec int64, processParams []ProcessParam) string {
	sb := new(strings.Builder)

	if bucket.BucketType == 1 {
		sb.WriteString("//")
		sb.WriteString(bucket.Domain)
		sb.WriteString("/")
		sb.WriteString(fileKey)

		if len(processParams) > 0 {
			sb.WriteString("?x-oss-process=")

			for i, porcessParam := range processParams {
				sb.WriteString(porcessParam.Name)

				if len(porcessParam.Params) > 0 {
					sb.WriteString(",")

					for i, param := range porcessParam.Params {
						sb.WriteString(param)

						if i < len(porcessParam.Params)-1 {
							sb.WriteString(",")
						}
					}
				}

				if i < len(processParams)-1 {
					sb.WriteString("/")
				}
			}
		} else if len(processParams) == 0 && bucket.ProcessConfig != "" {
			processConfig := &ProcessConfig{}

			if err := json.Unmarshal(([]byte)(bucket.ProcessConfig), &processConfig); err != nil {
				log.Logger.Error("解析ProcessConfig失败", zap.Any("ProcessConfig", bucket.ProcessConfig), zap.Error(err))
			} else if processConfig.Expression != "" {
				if expression, err := govaluate.NewEvaluableExpressionWithFunctions(processConfig.Expression, functions); err != nil {
					log.Logger.Error("解析ProcessConfig.Expression失败", zap.Any("ProcessConfig.Expression", processConfig.Expression), zap.Error(err))
				} else {
					parameters := make(map[string]any)

					parameters["bucket"] = bucket.Name
					parameters["fileKey"] = fileKey
					parameters["fileType"] = filepath.Ext(fileKey)

					if result, err := expression.Evaluate(parameters); err != nil {
						log.Logger.Error("执行ProcessConfig.Expression失败", zap.Any("ProcessConfig.Expression", processConfig.Expression), zap.Error(err))
					} else {
						if result, ok := result.(bool); ok && result {
							if len(processConfig.ProcessParams) > 0 {
								sb.WriteString("?x-oss-process=")

								for i, porcessParam := range processConfig.ProcessParams {
									sb.WriteString(porcessParam.Name)

									if len(porcessParam.Params) > 0 {
										sb.WriteString(",")

										for i, param := range porcessParam.Params {
											sb.WriteString(param)

											if i < len(porcessParam.Params)-1 {
												sb.WriteString(",")
											}
										}
									}

									if i < len(processConfig.ProcessParams)-1 {
										sb.WriteString("/")
									}
								}
							}
						}
					}
				}
			}

		}
	} else if bucket.BucketType == 2 {
		ossBucket, _ := ossClient.Bucket(bucket.Name)

		options := make([]oss.Option, 0, len(processParams))

		if len(processParams) > 0 {
			processParamSb := new(strings.Builder)

			for i, porcessParam := range processParams {
				processParamSb.WriteString(porcessParam.Name)

				if len(porcessParam.Params) > 0 {
					processParamSb.WriteString(",")

					for i, param := range porcessParam.Params {
						processParamSb.WriteString(param)

						if i < len(porcessParam.Params)-1 {
							processParamSb.WriteString(",")
						}
					}
				}

				if i < len(processParams)-1 {
					processParamSb.WriteString("/")
				}
			}

			options = append(options, oss.Process(processParamSb.String()))
		} else if len(processParams) == 0 && bucket.ProcessConfig != "" {
			processConfig := &ProcessConfig{}

			if err := json.Unmarshal(([]byte)(bucket.ProcessConfig), &processConfig); err != nil {
				log.Logger.Error("解析ProcessConfig失败", zap.Any("ProcessConfig", bucket.ProcessConfig), zap.Error(err))
			} else if processConfig.Expression != "" {
				if expression, err := govaluate.NewEvaluableExpressionWithFunctions(processConfig.Expression, functions); err != nil {
					log.Logger.Error("解析ProcessConfig.Expression失败", zap.Any("ProcessConfig.Expression", processConfig.Expression), zap.Error(err))
				} else {
					parameters := make(map[string]any)

					parameters["bucket"] = bucket.Name
					parameters["fileKey"] = fileKey
					parameters["fileType"] = filepath.Ext(fileKey)

					if result, err := expression.Evaluate(parameters); err != nil {
						log.Logger.Error("执行ProcessConfig.Expression失败", zap.Any("ProcessConfig.Expression", processConfig.Expression), zap.Error(err))
					} else {
						if result, ok := result.(bool); ok && result {
							if len(processConfig.ProcessParams) > 0 {
								processParamSb := new(strings.Builder)

								for i, porcessParam := range processConfig.ProcessParams {
									processParamSb.WriteString(porcessParam.Name)

									if len(porcessParam.Params) > 0 {
										processParamSb.WriteString(",")

										for i, param := range porcessParam.Params {
											processParamSb.WriteString(param)

											if i < len(porcessParam.Params)-1 {
												processParamSb.WriteString(",")
											}
										}
									}

									if i < len(processConfig.ProcessParams)-1 {
										processParamSb.WriteString("/")
									}
								}

								options = append(options, oss.Process(processParamSb.String()))
							}
						}
					}
				}
			}
		}

		signedURL, err := ossBucket.SignURL(fileKey, oss.HTTPGet, expiredInSec, options...)

		fmt.Println(signedURL)

		if err == nil && signedURL != "" && strings.Contains(signedURL, "//") {
			str := signedURL[strings.Index(signedURL, "//"):]

			return strings.Replace(str, fmt.Sprintf("%s.%s", bucket.Name, app.Endpoint), bucket.Domain, 1)
		}

	}

	return sb.String()
}

func buildSearchFilesWhere(searchParam SearchFilePageParam) (string, []any) {
	sb := new(strings.Builder)
	params := make([]any, 0, len(searchParam.Ids)+len(searchParam.FileKeys)+2)

	sb.WriteString(" WHERE f.del_status=0 AND b.del_status=0 AND a.del_status=0")

	if len(searchParam.Ids) > 0 {
		sb.WriteString(" AND f.id IN(")

		for i, id := range searchParam.Ids {
			sb.WriteString("?")

			if i < len(searchParam.Ids)-1 {
				sb.WriteString(",")
			}

			params = append(params, id)
		}

		sb.WriteString(")")
	}

	if len(searchParam.FileKeys) > 0 {
		sb.WriteString(" AND f.file_key IN(")

		for i, fileKey := range searchParam.FileKeys {
			sb.WriteString("?")

			if i < len(searchParam.FileKeys)-1 {
				sb.WriteString(",")
			}

			params = append(params, fileKey)
		}

		sb.WriteString(")")
	}

	if len(searchParam.Buckets) > 0 {
		sb.WriteString(" AND b.name IN(")

		for i, bucket := range searchParam.Buckets {
			sb.WriteString("?")

			if i < len(searchParam.Buckets)-1 {
				sb.WriteString(",")
			}

			params = append(params, bucket)
		}

		sb.WriteString(")")
	}

	return sb.String(), params
}

func mergeFiles(engine *xorm.Engine, files *[]File, expiredInSec int64, processParams []ProcessParam) error {
	if len(*files) == 0 {
		return nil
	}

	bucketIdMap := make(map[int64]int64)
	bucketIds := make([]int64, 0)

	for _, file := range *files {
		if _, ok := bucketIdMap[file.BucketId]; !ok {
			bucketIdMap[file.BucketId] = file.BucketId
			bucketIds = append(bucketIds, file.BucketId)
		}
	}

	bucketMap, err1 := getBucketMap(engine, bucketIds)

	if err1 != nil {
		return err1
	}

	appIdMap := make(map[int64]int64)
	appIds := make([]int64, 0)

	for _, v := range bucketMap {
		if _, ok := appIdMap[v.AppId]; !ok {
			appIdMap[v.AppId] = v.AppId
			appIds = append(appIds, v.AppId)
		}
	}

	appMap, err2 := getAppMap(engine, appIds)

	if err2 != nil {
		return err2
	}

	for i, file := range *files {
		if bucketEntity, ok := bucketMap[file.BucketId]; ok {
			if appEntity, ok := appMap[bucketEntity.AppId]; ok {
				if ossClient, err := getOssClientByBucket(appEntity); err == nil {
					(*files)[i].Url = getUrl(ossClient, appEntity, bucketEntity, file.FileKey, expiredInSec, processParams)
				}
			}
		}
	}

	return nil
}

func getBucketMap(engine *xorm.Engine, ids []int64) (map[int64]repository.Bucket, error) {
	if entities, err := repository.BucketRepository.FindByIdIn(engine, ids); err != nil {
		log.Logger.Error("getBucketMap", zap.Error(err))

		return nil, err
	} else {
		mMap := make(map[int64]repository.Bucket)

		for _, entity := range entities {
			mMap[entity.Id] = entity
		}

		return mMap, nil
	}
}

func getAppMap(engine *xorm.Engine, ids []int64) (map[int64]repository.App, error) {
	if entities, err := repository.AppRepository.FindByIdIn(engine, ids); err != nil {
		return nil, err
	} else {
		mMap := make(map[int64]repository.App)

		for _, entity := range entities {
			mMap[entity.Id] = entity
		}

		return mMap, nil
	}
}

func ossUploadFile(bucket *oss.Bucket, fileKey string, file string) error {
	var err error
	retryCount := 0
	retryMax := 3

	if err = bucket.UploadFile(fileKey, file, 100*1024, oss.Routines(3), oss.Checkpoint(true, "")); err != nil {
		log.Logger.Error("bucket.UploadFile", zap.String("fileKey", fileKey), zap.String("file", file), zap.Error(err))

		for retryCount < retryMax && err != nil {
			if err = bucket.UploadFile(fileKey, file, 100*1024, oss.Routines(3), oss.Checkpoint(true, "")); err != nil {
				log.Logger.Error("bucket.UploadFile", zap.String("fileKey", fileKey), zap.String("file", file), zap.Error(err))
			}

			retryCount++
		}
	}

	return err
}

func CreateFileData(file File) ([]File, error) {

	ms := make([]File, 0)
	engine := GetDB()

	fileEntity := &repository.File{
		BucketId:       file.BucketId,
		FileKey:        file.FileKey,
		SourceFile:     file.SourceFile,
		SourceFileType: file.SourceFileType,
		SourceFileSize: file.SourceFileSize,
		SourceFileAttr: file.SourceFileAttr,
	}

	now := util.GetCurrentLocalDateTime()

	fileEntity.CreateTime = now
	fileEntity.UpdateTime = now

	if err := repository.FileRepository.Create(engine, fileEntity); err != nil || fileEntity.Id == 0 {
		return ms, nil
	}

	sb := new(strings.Builder)
	params := make([]any, 0, 1)

	sb.WriteString("SELECT f.*,b.name AS bucket_name,b.domain FROM file f JOIN bucket b ON f.bucket_id=b.id JOIN app a ON b.app_id=a.id")
	sb.WriteString(" WHERE f.del_status=0 AND b.del_status=0 AND a.del_status=0")

	sb.WriteString(" AND f.file_key = ")

	sb.WriteString("?")

	params = append(params, file.FileKey)

	if err := engine.SQL(sb.String(), params...).Find(&ms); err != nil {
		return nil, err
	}

	expiredInSec := int64(60 * 60)

	if file.BucketId > 0 && file.BucketId == 3 {
		expiredInSec = 3600
	}

	if err := mergeFiles(engine, &ms, expiredInSec, nil); err != nil {
		return nil, err
	}

	return ms, nil
}

func BatchCreateFile(files []File) ([]File, error) {

	ms := make([]File, 0)
	engine := GetDB()

	for _, file := range files {

		msd := make([]File, 0)

		fileEntity := &repository.File{
			BucketId:       file.BucketId,
			FileKey:        file.FileKey,
			SourceFile:     file.SourceFile,
			SourceFileType: file.SourceFileType,
			SourceFileSize: file.SourceFileSize,
			SourceFileAttr: file.SourceFileAttr,
		}

		now := util.GetCurrentLocalDateTime()

		fileEntity.CreateTime = now
		fileEntity.UpdateTime = now

		if err := repository.FileRepository.Create(engine, fileEntity); err != nil || fileEntity.Id == 0 {
			return msd, nil
		}

		sb := new(strings.Builder)
		params := make([]any, 0, 1)

		sb.WriteString("SELECT f.*,b.name AS bucket_name,b.domain FROM file f JOIN bucket b ON f.bucket_id=b.id JOIN app a ON b.app_id=a.id")
		sb.WriteString(" WHERE f.del_status=0 AND b.del_status=0 AND a.del_status=0")

		sb.WriteString(" AND f.file_key = ")

		sb.WriteString("? ORDER BY create_time DESC LIMIt 1")

		params = append(params, file.FileKey)

		if err := engine.SQL(sb.String(), params...).Find(&msd); err != nil {
			return msd, err
		}

		expiredInSec := int64(60 * 60)

		if file.BucketId > 0 && file.BucketId == 3 {
			expiredInSec = 3600
		}

		if err := mergeFiles(engine, &msd, expiredInSec, nil); err != nil {
			return msd, err
		}

		ms = append(ms, msd[0])

	}

	return ms, nil
}
