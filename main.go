package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"path/filepath"
)

type DownloadLink struct {
	Url string
	Wallpaper string
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
	resp := getData(url)
	f, err := os.Create(path + fileName + ".jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.Write(*resp)
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
	path := filepath.Dir(os.Args[0]) + pathSeparator + "Wallpaper" + pathSeparator
	fmt.Println(path)
	// 1920x1080图片保存路径（无水印）
	path_1920x1080 := path + "1920x1080" + pathSeparator
	// 1920x1200图片保存路径（有水印）
	path_1920x1200 := path + "1920x1200" + pathSeparator
	// 判断当前目录是否存在Wallpaper文件夹，没有则创建
	_, err := os.Stat(path)
	if err != nil {
		os.MkdirAll(path_1920x1080, 0644)
		os.MkdirAll(path_1920x1200, 0644)
	}
	// 遍历壁纸链接
	for _, v := range loli.MediaContents {
		// 下载1920x1080
		url := "https://cn.bing.com" + v.ImageContent.Image.Url
		fileName := v.ImageContent.Title
		go saveImage(url, fileName, path_1920x1080, ch)
		// 下载1920x1200
		url = "https://cn.bing.com" + v.ImageContent.Image.Wallpaper
		go saveImage(url, fileName, path_1920x1200, ch)
	}
	for i := 0; i < len(loli.MediaContents) * 2; i++ {
		<-ch
	}
}