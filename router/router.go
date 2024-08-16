package router

import (
	"github.com/easynet-cn/file-service/controller"
	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/winter"
)

var (
	GinApplication = winter.NewApplication(object.Nacos)
)

func RunApplication() {
	GinApplication.Run(
		object.Nacos.Init,
		object.Database.Init,
		InitLogger,
		InitRouter)
}

func InitLogger() {
	log.Logger = winter.NewLogger(object.Nacos.GetConfig())
}

func InitRouter() {
	server := GinApplication.GetEngine()

	winter.RegisterDefaultMiddleware(server, &winter.SystemMiddleware{
		Logger:     log.Logger,
		Config:     object.Nacos.GetConfig(),
		Version:    object.Version,
		SyncDBFunc: object.SyncDB,
	})

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
}
