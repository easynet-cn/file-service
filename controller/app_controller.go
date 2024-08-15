package controller

import (
	"net/http"
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/winter"
	"github.com/gin-gonic/gin"
)

type appController struct{}

var AppController = &appController{}

func (c *appController) SearchPage(ctx *gin.Context) {
	searchParam := &winter.PageParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if pageResult, err := object.SearchApps(*searchParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, pageResult)
	}
}

func (c *appController) Create(ctx *gin.Context) {
	m := &object.App{}

	if err := ctx.BindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if app, err := object.CreateApp(*m); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, app)
	}
}

func (c *appController) Update(ctx *gin.Context) {
	m := &object.App{}

	if err := ctx.BindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else {
		if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err == nil {
			m.Id = id
		}

		if entity, err := object.UpdateApp(*m); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
		} else {
			ctx.JSON(http.StatusOK, entity)
		}
	}

}

func (c *appController) Delete(ctx *gin.Context) {
	if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
	} else if affected, err := object.DeleteAppById(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, &winter.RestResult{Status: 200, Data: affected > 0})
	}
}
