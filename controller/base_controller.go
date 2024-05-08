package controller

import (
	"strconv"

	"github.com/easynet-cn/file-service/object"
	"github.com/gin-gonic/gin"
)

const gatewayHeaderName = "Forwarded"

type ExtContext struct {
	context *gin.Context
}

func (ctx *ExtContext) JSON(code int, obj any) {
	var result = obj

	if ctx.context.Request.Header.Get(gatewayHeaderName) != "" {
		if code != 200 {
			result = &object.WebResult{Code: strconv.Itoa(code), Msg: obj.(string)}
		} else if pageResult, ok := obj.(*object.PageResult); ok {
			result = &object.WebResult{Code: "200", Data: pageResult.Data, Total: pageResult.Total}
		} else if pageResult, ok := obj.(object.PageResult); ok {
			result = &object.WebResult{Code: "200", Data: pageResult.Data, Total: pageResult.Total}
		} else {
			result = &object.WebResult{Code: "200", Data: obj}
		}
	} else {
		if code != 200 {
			result = &object.RestResult{Status: code, Message: obj.(string)}
		}
	}

	ctx.context.JSON(code, result)
}
