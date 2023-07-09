package scraper

import (
	"fmt"
	"testing"
)

func TestBangumi_DoReq(t *testing.T) {
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
	data, err := BangumiScraper.DoReq("POST", "https://api.bgm.tv/v0/search/subjects", &body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func TestBangumi_GetItem(t *testing.T) {
	item, err := BangumiScraper.GetItem("https://api.bgm.tv/v0/subjects/432980")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", item)
}
