package main

import (
	"fmt"
	"net/http"

	"github.com/easynet-cn/file-service/log"
	"github.com/easynet-cn/file-service/object"
	"github.com/easynet-cn/file-service/router"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func main() {
	object.InitNacos()
	log.InitLogger(object.Config)
	object.InitDB(object.Config)

	newRouter := router.NewRouter(object.Config)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", object.Config.GetInt("server.port")),
		Handler:      newRouter,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}

	log.Logger.Info("Service started successfully")

	if err := s.ListenAndServe(); err != nil {
		log.Logger.Error("Service startup failed", zap.Error(err))
	}
}
