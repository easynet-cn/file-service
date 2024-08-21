package controller

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/winter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fileController struct{}

var FileController = &fileController{}

func (c *fileController) Search(ctx *gin.Context) {
	searchParam := &object.SearchFileParam{}

	if err := ctx.ShouldBind(&searchParam); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if ms, err := object.SearchFiles(*searchParam); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, ms)
	}
}

func (c *fileController) SearchPage(ctx *gin.Context) {
	searchParam := &object.SearchFilePageParam{}

	if err := ctx.ShouldBind(&searchParam); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if pageResult, err := object.SearchPageFiles(*searchParam); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, pageResult)
	}
}

func (c *fileController) GetUploadToken(ctx *gin.Context) {
	m := &object.OssUploadFile{}

	if err := ctx.ShouldBind(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if uploadToken, err := object.GetUploadToken(*m); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, uploadToken)
	}
}

func (c *fileController) Upload(ctx *gin.Context) {
	if file, err := ctx.FormFile("file"); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if file == nil {
		winter.RenderBadRequestResult(ctx, errors.New("上传文件不能为空"))
	} else {
		fileExt := strings.TrimPrefix(path.Ext(file.Filename), ".")
		tempFile := path.Join(os.TempDir(), uuid.NewString()+path.Ext(file.Filename))

		if err := ctx.SaveUploadedFile(file, tempFile); err != nil {
			log.Logger.Error("保存上传文件失败", zap.Any("file", file), zap.Any("tempFile", tempFile), zap.Error(err))

			winter.RenderInternalServerErrorResult(ctx, err)
		} else {
			m := &object.OssUploadFile{}

			if err := ctx.ShouldBind(&m); err != nil {
				winter.RenderBadRequestResult(ctx, err)
			} else {
				m.SourceFile = file.Filename
				m.SourceFileSize = file.Size
				m.SourceFileType = fileExt

				if m.ProcessParamsStr != "" {
					processParams := make([]object.ProcessParam, 0)

					if err := json.Unmarshal([]byte(m.ProcessParamsStr), &processParams); err != nil {
						log.Logger.Error("解析ProcessParamsStr失败", zap.String("ProcessParamsStr", m.ProcessParamsStr), zap.Error(err))
					} else {
						m.ProcessParams = processParams
					}
				}

				if file, err := object.UploadFile(*m, tempFile); err != nil {
					log.Logger.Error("上传文件失败", zap.Any("uploadFile", m), zap.String("tempFile", tempFile), zap.Error(err))

					winter.RenderInternalServerErrorResult(ctx, err)
				} else {
					winter.RenderOkResult(ctx, file)
				}
			}
		}
	}
}

func (c *fileController) UploadBase64(ctx *gin.Context) {
	m := &object.OssUploadBase64{}

	if err := ctx.BindJSON(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if m.Data == "" {
		winter.RenderBadRequestResult(ctx, errors.New("上传文件不能为空"))
	} else if bytes, err := base64.StdEncoding.DecodeString(m.Data); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else {
		fileExt := strings.TrimPrefix(path.Ext(m.SourceFile), ".")
		tempFile := path.Join(os.TempDir(), uuid.NewString()+path.Ext(m.SourceFile))

		if err = os.MkdirAll(filepath.Dir(tempFile), 0750); err != nil {
			log.Logger.Error("创建临时文件夹失败", zap.Any("file", m), zap.Any("tempFile", tempFile), zap.Error(err))

			winter.RenderInternalServerErrorResult(ctx, err)
		}

		if err := os.WriteFile(tempFile, bytes, 0666); err != nil {
			log.Logger.Error("保存上传文件失败", zap.Any("file", m), zap.Any("tempFile", tempFile), zap.Error(err))

			winter.RenderInternalServerErrorResult(ctx, err)
		} else {
			if m.SourceFileType == "" {
				m.SourceFileType = fileExt
			}

			if m.ProcessParamsStr != "" {
				processParams := make([]object.ProcessParam, 0)

				if err := json.Unmarshal([]byte(m.ProcessParamsStr), &processParams); err != nil {
					log.Logger.Error("解析ProcessParamsStr失败", zap.String("ProcessParamsStr", m.ProcessParamsStr), zap.Error(err))
				} else {
					m.ProcessParams = processParams
				}
			}

			if file, err := object.UploadFile(m.OssUploadFile, tempFile); err != nil {
				log.Logger.Error("上传文件失败", zap.Any("uploadFile", m), zap.String("tempFile", tempFile), zap.Error(err))

				winter.RenderInternalServerErrorResult(ctx, err)
			} else {
				winter.RenderOkResult(ctx, file)
			}
		}
	}
}

func (c *fileController) Create(ctx *gin.Context) {
	m := &object.File{}

	if err := ctx.ShouldBindJSON(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if files, err := object.CreateFileData(*m); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, files)
	}
}

func (c *fileController) CreateBatch(ctx *gin.Context) {
	ms := make([]object.File, 0)

	if err := ctx.ShouldBindJSON(&ms); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if count, err := object.BatchCreateFile(ms); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, winter.NewRestResult(http.StatusOK, "200", count, ""))
	}
}
