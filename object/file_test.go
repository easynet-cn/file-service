package object

import (
	"fmt"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func Test_getUrl(t *testing.T) {
	options := make([]oss.Option, 0)

	options = append(options, oss.Process("image/watermark,image_VHVsaXBzLmpwZw,g_west,x_10,y_10"))
	options = append(options, oss.Process("image/resize,w_300"))

	fmt.Println(oss.GetRawParams(options))
}
