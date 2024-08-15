package object

import (
	"runtime"

	"github.com/easynet-cn/winter"
)

var (
	Nacos = winter.NewNacos(map[string]string{"goVersion": runtime.Version(), "version": Version})
)
