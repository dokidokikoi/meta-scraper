package scraper

import (
	"fmt"
	"testing"
)

func TestTwoDFan_GetItem(t *testing.T) {
	item, err := TwoDFanScraper.GetItem("https://2dfan.com/subjects/4566")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", item)
}
