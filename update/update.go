package update

import (
	"bufio"
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
	"io"
	"net/http"
	"os"
	"time"
)

type FileList struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FileResource struct {
	Data      string `json:"data"`
	ErrCode   string `json:"errCode"`
	ErrMsg    string `json:"errMsg"`
	FileId    int    `json:"fileId"`
	ReturnUrl string `json:"returnUrl"`
	Status    string `json:"status"`
}

// DownloadFile 下载音频文件
func DownloadFile(page int) (fileAddress []string, err error) {
	ctx := gctx.New()
	page = page + 1
	fileAddress, err = GetFileResource(ctx, page)
	if err != nil {
		return nil, err
	}
	fileAddress, err = DownloadFile(page)
	fileAddress = append(fileAddress, fileAddress...)
	if err != nil {
		return nil, err
	}

	return
}

// GetFileId 获取文件名和id
func GetFileId(ctx context.Context, pageLimit string) (fileList []*FileList, err error) {
	url := nextUrl + resourcesId + "/" + pageLimit + "/next"
	c := g.Client()
	c.SetHeader("Content-Type", "application/json; charset=UTF-8")
	res, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic(err)
	}
	var names []string
	var ids []string
	doc.Find(".section li .column1").Each(func(i int, s *goquery.Selection) {
		name := s.ChildrenFiltered(`.column1 span`).Text()
		names = append(names, name)
	})
	doc.Find(".section li .column1 input[name=sectionid]").Each(func(i int, s *goquery.Selection) {
		nodes := s.Nodes
		for _, node := range nodes {
			attrs := node.Attr
			for _, attr := range attrs {
				if attr.Key == "value" {
					//val := gstr.SubStr(attr.Val, 8, 8)
					ids = append(ids, attr.Val)
				}
			}
		}
	})
	var fileArray garray.Array
	for i := 0; i < len(ids); i++ {
		fileArray.Append(g.Map{
			"id":   ids[i],
			"name": names[i],
		})
	}
	err = gconv.Struct(fileArray.Slice(), &fileList)
	return
}

// GetFileResource 获取文件资源
func GetFileResource(ctx context.Context, pageSize int) (fileAddress []string, err error) {
	pageLimit := (pageSize-1)*10 + 1
	fileList, err := GetFileId(ctx, gconv.String(pageLimit))
	if err != nil || len(fileList) == 0 {
		return nil, err
	}
	var fileResource *FileResource
	fileUrl := fileUrl + resourcesId + "/"
	for _, file := range fileList {
		time.Sleep(sleepTime)
		var address string
		url := fileUrl + file.Id
		c := g.Client()
		c.SetHeader("Content-Type", "application/json; charset=UTF-8")
		c.SetHeader("Accept", "*/*")
		c.SetHeader("Cookie", cookie)
		c.SetHeader("Referer", "https://www.lrts.me/playlist")
		c.SetHeader("sec-ch-ua", "Microsoft Edge\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24")
		c.SetHeader("sec-ch-ua-mobile", "?0")
		res, _ := c.Get(ctx, url)
		result := res.ReadAll()
		err = gconv.Struct(gconv.String(result), &fileResource)
		if err != nil {
			return nil, err
		}
		address, err = GetFileAndSave(fileResource.Data, file)
		if err != nil {
			return nil, err
		}
		fileAddress = append(fileAddress, address)
		time.Sleep(sleepTime)
	}
	return
}

// GetFileAndSave 保存文件信息
func GetFileAndSave(fileUrl string, fileList *FileList) (url string, err error) {
	time.Sleep(5)
	res, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	reader := bufio.NewReaderSize(res.Body, 32*1024)
	file, err := os.Create(filePath + fileName + fileList.Name + ".mp3")
	if err != nil {
		return "", err
	}
	writer := bufio.NewWriter(file)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return "", err
	}
	url = filePath + fileName + fileList.Name + ".mp3"
	return
}
