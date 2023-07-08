package scraper

import (
	"fmt"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	data, err := GetChuScraper.DoReq("https://www.getchu.com/soft.phtml?id=1219845c")
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.OpenFile("./html/getchu.html", os.O_CREATE|os.O_WRONLY, 0774)
	if err != nil {
		fmt.Println("open file error", err)
		return
	}
	_, err = f.Write(data)
	if err != nil {
		fmt.Println("write file error", err)
		return
	}
	f.Close()
	fmt.Println(string(data))
}

func TestGetChu_GetItem(t *testing.T) {
	item, err := GetChuScraper.GetItem("https://www.getchu.com/soft.phtml?id=1219845c&gc=gc")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", item)
}
