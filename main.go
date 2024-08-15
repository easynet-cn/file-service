package main

import (
	"github.com/easynet-cn/file-service/router"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	router.RunApplication()
}
