package scraper

import (
	"fmt"
	"os"
	"testing"
)

func TestGGBases_DO(t *testing.T) {
	data, err := GGBasesScraper.DoReq("")
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.OpenFile("./html/ggbases.html", os.O_CREATE|os.O_WRONLY, 0774)
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
}

func TestGGBases_GetItem(t *testing.T) {
	item, err := GGBasesScraper.GetItem("https://ggbases.dlgal.com/view.so?id=119583")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", item)
}
