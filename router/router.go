package router

import (
	"fmt"
	"net/http"
	"runtime"

	ginzap "github.com/gin-contrib/zap"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/easynet-cn/file-service/controller"
	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/object"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func NewRouter(viper *viper.Viper) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	server := gin.Default()

	server.Use(ginzap.Ginzap(log.Logger, viper.GetString("logging.date-time-format"), false))
	server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(Recovery)

	server.GET("/system/stats", Stats)
	server.GET("/system/version", Version)
	server.GET("/db/sync", Sync)

	apiGroup := server.Group("/v1/")

	apiGroup.POST("/apps/search/page", controller.AppController.SearchPage) //应用分页查询
	apiGroup.POST("/apps", controller.AppController.Create)                 //创建应用
	apiGroup.PUT("/apps/:id", controller.AppController.Update)              //更新应用
	apiGroup.DELETE("/apps/:id", controller.AppController.Delete)           //删除应用

	apiGroup.POST("/buckets/search/page", controller.BucketController.SearchPage) //存储空间分页查询
	apiGroup.POST("/buckets", controller.BucketController.Create)                 //创建存储空间
	apiGroup.PUT("/buckets/:id", controller.BucketController.Update)              //更新存储空间
	apiGroup.DELETE("/buckets/:id", controller.BucketController.Delete)           //删除存储空间

	apiGroup.POST("/files/search", controller.FileController.Search)               //文件查询
	apiGroup.POST("/files/search/page", controller.FileController.SearchPage)      //文件分页查询
	apiGroup.POST("/files/upload/token", controller.FileController.GetUploadToken) //获取上传凭证
	apiGroup.POST("/files/upload", controller.FileController.Upload)               //上传文件
	apiGroup.POST("/files/upload/base64", controller.FileController.UploadBase64)  //上传Base64文件
	apiGroup.POST("/files", controller.FileController.Create)                      //创建文件数据
	apiGroup.POST("/files/batch", controller.FileController.CreateBatch)           //批量创建文件数据

	return server
}

func Recovery(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Error("router", zap.Error(fmt.Errorf("%v", r)))

			ctx.JSON(http.StatusOK, object.RestResult{Status: 500, Code: "500", Message: "系统内部错误"})
		}
	}()

	ctx.Next()
}

func Stats(ctx *gin.Context) {
	stats := &runtime.MemStats{}

	runtime.ReadMemStats(stats)

	ctx.JSON(http.StatusOK, stats)
}

func Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, object.NewSystemVersion())
}

func Sync(ctx *gin.Context) {
	if err := object.SyncDB(); err != nil {
		ctx.JSON(http.StatusInternalServerError, object.RestResult{Status: 500, Code: "500", Message: err.Error()})
	} else {
		ctx.JSON(http.StatusOK, object.RestResult{Status: 200, Code: "200", Message: "同步成功"})
	}
}
