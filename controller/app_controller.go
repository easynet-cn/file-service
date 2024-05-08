package controller

import (
	"net/http"
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/gin-gonic/gin"
)

type appController struct{}

var AppController = &appController{}

func (c *appController) SearchPage(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	searchParam := &object.PageParam{}

	if err := ctx.BindJSON(&searchParam); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if pageResult, err := object.SearchApps(*searchParam); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, pageResult)
	}
}

func (c *appController) Create(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.App{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if app, err := object.CreateApp(*m); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		context.JSON(http.StatusOK, app)
	}
}

func (c *appController) Update(ctx *gin.Context) {
	context := &ExtContext{context: ctx}
	m := &object.App{}

	if err := ctx.BindJSON(&m); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else {
		if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err == nil {
			m.Id = id
		}

		if entity, err := object.UpdateApp(*m); err != nil {
			context.JSON(http.StatusInternalServerError, err.Error())
		} else {
			context.JSON(http.StatusOK, entity)
		}
	}

}

func (c *appController) Delete(ctx *gin.Context) {
	context := &ExtContext{context: ctx}

	if id, err := strconv.ParseInt(ctx.Param("id"), 10, 64); err != nil {
		context.JSON(http.StatusBadRequest, err.Error())
	} else if affected, err := object.DeleteAppById(id); err != nil {
		context.JSON(http.StatusInternalServerError, err.Error())
	} else {
		if ctx.Request.Header.Get(gatewayHeaderName) != "" {
			context.JSON(http.StatusOK, affected > 0)
		} else {
			context.JSON(http.StatusOK, &object.RestResult{Status: 200, Data: affected > 0})
		}
	}
}
