package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

type DownloadLink struct {
	Url string
}

type Images struct {
	Image DownloadLink
	Title string
}

type ImageContentInfo struct {
	ImageContent Images
}

type Loli struct {
	MediaContents []ImageContentInfo
}

func getData(url string) *[]byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("无法访问Bing！")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return &body
}

// 保存图片
func saveImage(url string, fileName string, path string, ch chan int) {
	if _, err := os.Stat(path + fileName); err == nil {
		fmt.Printf("%s 已下载，正在跳过...\n", fileName)
		ch <- 1
		return
	}
	fmt.Printf("正在下载-->%s\n", fileName)
	resp := getData(url)
	f, err := os.Create(path + fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(*resp)
	fmt.Println("下载成功！")
	ch <- 1
}

func main() {
	var loli Loli
	// 设置3个下载线程
	ch := make(chan int, 3)
	loli.MediaContents = make([]ImageContentInfo, 1)
	// 获取Bing壁纸json数据
	resp := getData("https://cn.bing.com/hp/api/model")
	// 解析json数据
	json.Unmarshal(*resp, &loli)
	// 路径分隔符
	pathSeparator := filepath.FromSlash("/")
	// 获取当前运行目录的Wallpaper文件夹
	path := filepath.Dir(os.Args[0]) + pathSeparator + "wallpaper" + pathSeparator
	_, err := os.Stat(path)
	if err != nil {
		os.Mkdir(path, 0775)
	}
	re := regexp.MustCompile("id=(.+?)_1920")
	// 遍历壁纸链接
	for _, v := range loli.MediaContents {
		// 下载1920x1080
		url := "https://cn.bing.com" + v.ImageContent.Image.Url
		imageId := re.FindStringSubmatch(url)
		fileName := v.ImageContent.Title
		go saveImage(url, fileName + "_1920x1080.jpg", path, ch)
		// 下载4K
		url = "https://www.bingimg.cn/down/uhd/"+ imageId[1] +"_UHD.jpg"
		go saveImage(url, fileName + "_UHD.jpg", path, ch)
	}
	for i := 0; i < len(loli.MediaContents) * 2; i++ {
		<-ch
	}
}