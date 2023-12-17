package main

import (
	"fmt"
	"os"
	"scraper/scraper"
	"scraper/tools"
)

func main() {
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
	headers["Referer"] = "https://2dfan.com/"
	headers["Accept-Language"] = "zh-CN,zh;q=0.9"
	data, _, err := tools.MakeRequest(
		"GET",
		"https://img.achost.top/uploads/subjects/packages/thumb_f263dcea44c2af791fbb01a0002712ad.jpg",
		scraper.BangumiScraper.Proxy, nil, headers, nil)
	if err != nil {
		fmt.Println("write file error", err)
		return
	}
	f, err := os.OpenFile("img.jpg", os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(data)
	if err != nil {
		fmt.Println("write file error", err)
		return
	}
	f.Close()
}
