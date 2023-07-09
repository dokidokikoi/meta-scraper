package main

import (
	"fmt"
	"scraper/scraper"
)

func main() {
	type Filter struct {
		Nfsw bool `json:"nfsw"`
	}
	body := struct {
		Keyword string `json:"keyword"`
		Sort    string `json:"sort"`
		Filter  Filter `json:"filter"`
	}{
		Keyword: "光装剣姫アークブレイバー 魔族篇胞",
		Sort:    "rank",
		Filter:  Filter{Nfsw: true},
	}
	data, err := scraper.BangumiScraper.DoReq("POST", "https://api.bgm.tv/v0/search/subjects", &body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
