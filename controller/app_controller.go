package controller

import (
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/winter"
	"github.com/gin-gonic/gin"
)

type appController struct{}

var AppController = &appController{}

func (c *appController) SearchPage(ctx *gin.Context) {
	searchParam := &winter.PageParam{}

	if err := ctx.ShouldBind(&searchParam); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if pageResult, err := object.SearchApps(*searchParam); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, pageResult)
	}
}

func (c *appController) Create(ctx *gin.Context) {
	m := &object.App{}

	if err := ctx.ShouldBind(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if app, err := object.CreateApp(*m); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderOkResult(ctx, app)
	}
}

func (c *appController) Update(ctx *gin.Context) {
	m := &object.App{}

	if err := ctx.ShouldBind(&m); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else {
		if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err == nil {
			m.Id = id
		}

		if entity, err := object.UpdateApp(*m); err != nil {
			winter.RenderInternalServerErrorResult(ctx, err)
		} else {
			winter.RenderOkResult(ctx, entity)
		}
	}
}

func (c *appController) Delete(ctx *gin.Context) {
	if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err != nil {
		winter.RenderBadRequestResult(ctx, err)
	} else if affected, err := object.DeleteAppById(id); err != nil {
		winter.RenderInternalServerErrorResult(ctx, err)
	} else {
		winter.RenderSuccessResult(ctx, &winter.RestResult{Status: 200, Data: affected > 0})
	}
}
