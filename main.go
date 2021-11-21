package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"strconv"
	"time"
)

type ImageInfo struct {
	Id string
	Title string
	Date string
}

// 历史图片下载
type HistoryImage struct {
	Image []ImageInfo
}

type DownloadLink struct {
	Url string
}

type Images struct {
	Image DownloadLink
	Title string
}

type ImageContentInfo struct {
	ImageContent Images
	Ssd string
}

// 每日图片下载
type Loli struct {
	MediaContents []ImageContentInfo
}

func getData(url string) *[]byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("链接请求失败！")
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return &body
}

// 下载计数
var count int

// 保存图片
func saveImage(Id string, imageType *[]string, title string, path string, blacklist *[]string) {
	// 过滤文件名特殊符号
	fileName := title
	for _, k := range *blacklist {
		if index := strings.Index(title, k); index != -1 {
			fileName = strings.Replace(fileName, k, "", -1)
		}
	}
	for _, v := range *imageType {
		filepath := path + fileName + "_" + v + ".jpg"
		if _, err := os.Stat(filepath); err == nil {
			fmt.Printf("%s 已下载，正在跳过...\n", fileName)
			continue
		}
		count++
		fmt.Printf("%d %s ", count, fileName + "_" + v + ".jpg")
		url := "https://cn.bing.com/th?id=" + Id + "_" + v + ".jpg&rf=LaDigue_" + v + ".jpg"
		resp := getData(url)
		// 判断图片大小是否为0
		if len(*resp) == 0 || *resp == nil {
			fmt.Println("下载失败，图片大小为0，已跳过")
			continue
		}
		f, err := os.Create(filepath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer f.Close()
		f.Write(*resp)
		fmt.Println("下载成功！")
	}
}

// 历史壁纸下载
func (historyImage *HistoryImage) HistoryDownload(imageType *[]string, path string, blacklist *[]string) {
	re := regexp.MustCompile(`alt="(.+?)_1920x1080\.jpg"\r\n\s+title="(.+?)\s\(&copy[\s \S]+?class="date">(.+?)</p>`)
	// 取出最后一页
	respPageMax := getData("http://bing.richex.cn")
	rePageMax := regexp.MustCompile(`\.\.\.</span></li><li>.+?>(\d+)</a></li> <li>`)
	resPageMax := rePageMax.FindSubmatch(*respPageMax)
	pageMax, _ := strconv.Atoi(string(resPageMax[1]))
	// 获取图片Id和标题
	for i := 1; i <= pageMax; i++ {
		fmt.Printf("正在爬取第%d页\n", i)
		resp := getData("http://bing.richex.cn/?page=" + fmt.Sprintf("%d", i))
		res := re.FindAllSubmatch(*resp, -1)
		for _, v := range res {
			if historyImage.Image[0].Id == "" {
				historyImage.Image[0].Id = string(v[1])
				historyImage.Image[0].Title = string(v[2])
				historyImage.Image[0].Date = string(v[3])
				continue
			}
			historyImage.Image = append(historyImage.Image, ImageInfo{string(v[1]), string(v[2]), string(v[3])})
		}
		time.Sleep(1 * time.Second)
	}
	// 开始下载历史图片
	for _, v := range historyImage.Image {
		saveImage(v.Id, imageType, v.Title + v.Date, path, blacklist)
	}
}

func main() {
	var loli Loli
	var historyImage HistoryImage
	loli.MediaContents = make([]ImageContentInfo, 1)
	historyImage.Image = make([]ImageInfo, 1)
	// 文件命名特殊符号过滤
	blacklist := []string{"\\u0026ldquo;", "\\u0026rdquo;", "\\", "/", ":", "*", "?", "<", ">", "|"}
	// 分辨率
	imageType := []string{"1920x1080", "UHD"}
	// 路径分隔符
	pathSeparator := filepath.FromSlash("/")
	// 获取当前运行目录的Wallpaper文件夹
	path := filepath.Dir(os.Args[0]) + pathSeparator + "wallpaper" + pathSeparator
	_, err := os.Stat(path)
	if err != nil {
		os.Mkdir(path, 0775)
	}
	// 根据启动参数长度来决定是否下载历史图片
	if len(os.Args) == 2 && os.Args[1] == "-loli" {
		historyImage.HistoryDownload(&imageType, path, &blacklist)
		return
	}
	// 获取Bing壁纸json数据
	resp := getData("https://cn.bing.com/hp/api/model")
	// 解析json数据
	json.Unmarshal(*resp, &loli)
	re := regexp.MustCompile("id=(.+?)_1920")
	// 遍历壁纸链接
	for _, v := range loli.MediaContents {
		url := "https://cn.bing.com" + v.ImageContent.Image.Url
		imageId := re.FindStringSubmatch(url)
		date := fmt.Sprintf("%s-%s-%s", v.Ssd[:4], v.Ssd[4:6], v.Ssd[6:8])
		saveImage(imageId[1], &imageType, v.ImageContent.Title + "_" + date, path, &blacklist)
	}
}