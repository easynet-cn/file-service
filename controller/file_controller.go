package controller

import (
	"encoding/base64"
	"encoding/json"
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

	if err := ctx.BindJSON(&searchParam); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if ms, err := object.SearchFiles(*searchParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, ms)
	}
}

func (c *fileController) SearchPage(ctx *gin.Context) {
	searchParam := &object.SearchFilePageParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if pageResult, err := object.SearchPageFiles(*searchParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, pageResult)
	}
}

func (c *fileController) GetUploadToken(ctx *gin.Context) {
	m := &object.OssUploadFile{}

	if err := ctx.BindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if uploadToken, err := object.GetUploadToken(*m); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, uploadToken)
	}
}

func (c *fileController) Upload(ctx *gin.Context) {
	if file, err := ctx.FormFile("file"); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if file == nil {
		ctx.JSON(http.StatusBadRequest, "上传文件不能为空")
	} else {
		fileExt := strings.TrimPrefix(path.Ext(file.Filename), ".")
		tempFile := path.Join(os.TempDir(), uuid.NewString()+path.Ext(file.Filename))

		if err := ctx.SaveUploadedFile(file, tempFile); err != nil {
			log.Logger.Error("保存上传文件失败", zap.Any("file", file), zap.Any("tempFile", tempFile), zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, err.Error())
		} else {
			m := &object.OssUploadFile{}

			if err := ctx.Bind(&m); err != nil {
				ctx.JSON(http.StatusBadRequest, err.Error())
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

					ctx.JSON(http.StatusInternalServerError, err.Error())
				} else {
					ctx.JSON(http.StatusOK, file)
				}
			}
		}
	}
}

func (c *fileController) UploadBase64(ctx *gin.Context) {
	m := &object.OssUploadBase64{}

	if err := ctx.BindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if m.Data == "" {
		ctx.JSON(http.StatusBadRequest, "上传文件不能为空")
	} else if bytes, err := base64.StdEncoding.DecodeString(m.Data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else {
		fileExt := strings.TrimPrefix(path.Ext(m.SourceFile), ".")
		tempFile := path.Join(os.TempDir(), uuid.NewString()+path.Ext(m.SourceFile))

		if err = os.MkdirAll(filepath.Dir(tempFile), 0750); err != nil {
			log.Logger.Error("创建临时文件夹失败", zap.Any("file", m), zap.Any("tempFile", tempFile), zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		if err := os.WriteFile(tempFile, bytes, 0666); err != nil {
			log.Logger.Error("保存上传文件失败", zap.Any("file", m), zap.Any("tempFile", tempFile), zap.Error(err))

			ctx.JSON(http.StatusInternalServerError, err.Error())
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

				ctx.JSON(http.StatusInternalServerError, err.Error())
			} else {
				ctx.JSON(http.StatusOK, file)
			}
		}
	}
}

func (c *fileController) Create(ctx *gin.Context) {
	m := &object.File{}

	if err := ctx.BindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if files, err := object.CreateFileData(*m); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, files)
	}
}

func (c *fileController) CreateBatch(ctx *gin.Context) {
	ms := make([]object.File, 0)

	if err := ctx.BindJSON(&ms); err != nil {
		ctx.JSON(http.StatusBadRequest, &winter.RestResult{Status: 400, Message: err.Error()})
	} else if count, err := object.BatchCreateFile(ms); err != nil {
		ctx.JSON(http.StatusInternalServerError, &winter.RestResult{Status: 500, Message: err.Error()})
	} else {
		ctx.JSON(http.StatusOK, count)
	}
}
