package controller

import (
	"net/http"
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/gin-gonic/gin"
)

type bucketController struct{}

var BucketController = &bucketController{}

func (c *bucketController) SearchPage(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	searchParam := &object.PageParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if pageResult, err := object.SearchBuckets(*searchParam); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, pageResult)
	}
}

func (c *bucketController) Create(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.Bucket{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if bucket, err := object.CreateBucket(*m); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, bucket)
	}
}

func (c *bucketController) Update(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.Bucket{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else {
		if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err == nil {
			m.Id = id
		}

		if entity, err := object.UpdateBucket(*m); err != nil {
			context.JSON(http.StatusInternalServerError, err.Error())
		} else {
			context.JSON(http.StatusOK, entity)
		}
	}

}

func (c *bucketController) Delete(ctx *gin.Context) {
	context := &ExtContext{context: ctx}

	if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if affected, err := object.DeleteBucketById(id); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		if ctx.Request.Header.Get(gatewayHeaderName) != "" {
			context.JSON(http.StatusOK, affected > 0)
		} else {
			context.JSON(http.StatusOK, &object.RestResult{Status: 200, Data: affected > 0})
		}
	}
}
