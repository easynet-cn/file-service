package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/object"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fileController struct{}

var FileController = &fileController{}

func (c *fileController) Search(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	searchParam := &object.SearchFileParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if ms, err := object.SearchFiles(*searchParam); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, ms)
	}
}

func (c *fileController) SearchPage(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	searchParam := &object.SearchFilePageParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if pageResult, err := object.SearchPageFiles(*searchParam); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, pageResult)
	}
}

func (c *fileController) GetUploadToken(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.OssUploadFile{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if uploadToken, err := object.GetUploadToken(*m); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, uploadToken)
	}
}

func (c *fileController) Upload(ctx *gin.Context) {
	context := &ExtContext{context: ctx}

	if file, err := ctx.FormFile("file"); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if file == nil {
		context.JSON(http.StatusBadRequest, "上传文件不能为空")
	} else {
		fileExt := strings.TrimPrefix(path.Ext(file.Filename), ".")
		tempFile := path.Join(os.TempDir(), uuid.NewString()+path.Ext(file.Filename))

		if err := ctx.SaveUploadedFile(file, tempFile); err != nil {
			log.Logger.Error("保存上传文件失败", zap.Any("file", file), zap.Any("tempFile", tempFile), zap.Error(err))

			context.JSON(http.StatusInternalServerError, err.Error())
		} else {
			m := &object.OssUploadFile{}

			if err := ctx.Bind(&m); err != nil {
				context.JSON(http.StatusBadRequest, err.Error())
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

					context.JSON(http.StatusInternalServerError, err.Error())
				} else {
					context.JSON(http.StatusOK, file)
				}
			}
		}
	}
}

func (c *fileController) Create(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.File{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if files, err := object.CreateFileData(*m); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, files)
	}
}

func (c *fileController) CreateBatch(ctx *gin.Context) {
	ms := make([]object.File, 0)

	if err := ctx.BindJSON(&ms); err != nil {
		ctx.JSON(http.StatusBadRequest, &object.RestResult{Status: 400, Message: err.Error()})
	} else if count, err := object.BatchCreateFile(ms); err != nil {
		ctx.JSON(http.StatusInternalServerError, &object.RestResult{Status: 500, Message: err.Error()})
	} else {
		ctx.JSON(http.StatusOK, count)
	}
}