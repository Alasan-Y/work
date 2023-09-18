package update_lrts

import (
	"github.com/gogf/gf/v2/frame/g"
	"update-lrts/update"
)

func main() {
	g.Dump("开始下载")
	page := 1
	fileAddress, err := update.DownloadFile(page)
	if err != nil {
		g.Dump("err ==>", err)
	}
	g.Dump("fileAddress ==>", fileAddress)

}
