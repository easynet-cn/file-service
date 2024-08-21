package controller

import (
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/winter"
	"github.com/gin-gonic/gin"
)

type bucketController struct{}

var BucketController = &bucketController{}

func (c *bucketController) SearchPage(ctx *gin.Context) {
	searchParam := &winter.PageParam{}

	if err := ctx.ShouldBind(&searchParam); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if pageResult, err := object.SearchBuckets(*searchParam); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, pageResult)
	}
}

func (c *bucketController) Create(ctx *gin.Context) {
	m := &object.Bucket{}

	if err := ctx.ShouldBind(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if bucket, err := object.CreateBucket(*m); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, bucket)
	}
}

func (c *bucketController) Update(ctx *gin.Context) {
	m := &object.Bucket{}

	if err := ctx.ShouldBind(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else {
		if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err == nil {
			m.Id = id
		}

		if entity, err := object.UpdateBucket(*m); err != nil {
			winter.RenderInternalServerErrorResult(ctx, err)
		} else {
			winter.RenderOkResult(ctx, entity)
		}
	}
}

func (c *bucketController) Delete(ctx *gin.Context) {
	if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if affected, err := object.DeleteBucketById(id); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, &winter.RestResult{Status: 200, Data: affected > 0})
	}
}
