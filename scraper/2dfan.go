package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"scraper/tools"
	"strings"
)

var (
	twoDFanDomain    = "https://2dfan.org/"
	twoDFanSearchUri = "https://2dfan.org/subjects/search?keyword=%s"
)

type TwoDFan struct {
	Proxy     string
	Domain    string
	SearchUri string
	Headers   map[string]string
}

var TwoDFanScraper *TwoDFan

func (tdf *TwoDFan) DoReq(method, uri string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	data, _, err = tools.MakeRequest(method, uri, tdf.Proxy, bytes.NewBuffer(data), tdf.Headers, nil)
	return data, err
}

func (tdf *TwoDFan) GetItem(uri string) (*Item, error) {
	data, err := tdf.DoReq("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	root, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	item := &Item{Origin: uri}
	// 获取名称
	item.Name, err = tdf.GetItemName(root)
	if err != nil {
		fmt.Println("获取名称失败 url:", uri, "err:", err)
	}
	// 获取品牌
	item.Brand, err = tdf.GetItemBrand(root)
	if err != nil {
		fmt.Println("获取品牌失败 url:", uri, "err:", err)
	}
	// 获取发售日
	item.ReleaseDate, err = tdf.GetItemReleaseDate(root)
	if err != nil {
		fmt.Println("获取发售日失败 url:", uri, "err:", err)
	}
	// 获取tag
	item.Tags, err = tdf.GetItemTags(root)
	if err != nil {
		fmt.Println("获取tag失败 url:", uri, "err:", err)
	}
	// 获取攻略
	u, err := url.Parse(uri)
	if err == nil {
		path := strings.Split(u.EscapedPath(), "/")
		id := path[len(path)-1]
		item.WalkThrough, err = tdf.GetItemWalkThrough(id)
		if err != nil {
			fmt.Println("获取攻略失败 url:", uri, "err:", err)
		}

		tdf.GetOtherInfo(id, root, item)
	} else {
		fmt.Println("获取攻略失败 url:", uri, "err:", err)
	}

	return item, nil
}

func (tdf *TwoDFan) GetItemName(node *goquery.Document) (string, error) {
	return node.Find("div.navbar h3").First().Text(), nil
}

func (tdf *TwoDFan) GetItemWalkThrough(id string) (string, error) {
	return fmt.Sprintf("https://2dfan.com/subjects/%s/walkthroughs", id), nil
}

func (tdf *TwoDFan) GetItemBrand(node *goquery.Document) (string, error) {
	brand := ""
	node.Find(`div[class="media-body control-group"] p.tags`).Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), "品牌") {
			brand = selection.Find("a").First().Text()
			return
		}
	})
	return brand, nil
}

func (tdf *TwoDFan) GetItemReleaseDate(node *goquery.Document) (string, error) {
	date := ""
	node.Find(`div[class="media-body control-group"] p.tags`).Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), "发售日期") {
			date = strings.Replace(selection.Text(), "发售日期：", "", 1)
			date = strings.TrimSpace(date)
			return
		}
	})
	return date, nil
}

func (tdf *TwoDFan) GetItemTags(node *goquery.Document) ([]Tag, error) {
	var tags []TagItem
	node.Find(`#sidebar div[class="block-content collapse in tags"] a`).Each(func(i int, selection *goquery.Selection) {
		tags = append(tags, TagItem{
			Identity: selection.Text(),
			Name:     selection.Text(),
		})

	})

	return []Tag{{Item: tags}}, nil
}

func (tdf *TwoDFan) GetOtherInfo(id string, node *goquery.Document, item *Item) {
	item.Preview = []string{}
	image, ok := node.Find(`div[class="block-content collapse in"] div.span8 div.media a img`).First().Attr("src")
	if ok {
		item.Preview = append(item.Preview, image)
	}

	node.Find(`#resources span`).Each(func(i int, selection *goquery.Selection) {
		if strings.Contains(selection.Text(), "介绍") {
			data, err := tdf.DoReq("Get", fmt.Sprintf("https://2dfan.com/topics/%s", id), nil)
			if err != nil {
				return
			}
			node, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
			if err != nil {
				return
			}

			story := strings.Builder{}
			information := strings.Builder{}

			f := func(node *goquery.Document) {
				node.Find("#topic-content").Children().Each(func(i int, selection *goquery.Selection) {
					if selection.Is("h4") {
						if strings.Contains(selection.Text(), "人物介绍") {
							item.Story = story.String()
						}
					}
					if item.Story == "" {

					}
					str := selection.Text()
					switch {
					case strings.Contains(str, "故事介绍"):
						for {
							selection = selection.Next()
							if selection.Is("h4") {
								item.Story = story.String()
								return
							}
							if selection.Is("img") {
								selection.Find("img").Each(func(i int, selection *goquery.Selection) {
									image, ok := selection.Attr("src")
									if ok {
										item.Preview = append(item.Preview, image)
									}
								})
								continue
							}
							story.WriteString(selection.Text())
							story.WriteByte('\n')
						}
					case strings.Contains(str, "人物介绍"):
						item.Character = []Character{}
						c := Character{}
						for {
							selection = selection.Next()
							if selection.Is("h4") {
								item.Story = story.String()
								return
							}

							if selection.Is("strong") {
								if !selection.Prev().Is("h4") {
									item.Character = append(item.Character, c)
									information.Reset()
								}

								c.Name = selection.Find("strong").First().Text()
								continue
							}
							if selection.Is("img") {
								c.Avatar, _ = selection.Find("img").First().Attr("src")

								selection = selection.Next()
								image, ok := selection.Find("img").First().Attr("src")
								if ok {
									c.Images = []string{image}
								}
								continue
							}

							information.WriteString(selection.Text())
							information.WriteByte('\n')
						}
					}
				})
			}

			f(node)
			if !node.Is("#content-pagination div.pagination") {
				return
			}
			n := node.Find("#content-pagination div.pagination ul li").Length() - 2
			for i = 1; i < n; i++ {
				data, err = tdf.DoReq("Get", fmt.Sprintf("https://2dfan.com/topics/%s/page/%d", id, i+1), nil)
				if err != nil {
					return
				}
				node, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
				if err != nil {
					return
				}
				f(node)
			}
			return
		}
	})
}

func init() {
	headers := make(map[string]string)
	headers["User-Agent"] = defaultUserAgent
	headers["Referer"] = twoDFanDomain
	headers["Accept-Language"] = "zh-CN,zh;q=0.9"
	TwoDFanScraper = &TwoDFan{
		Proxy:     defaultProxy,
		Domain:    twoDFanDomain,
		SearchUri: twoDFanSearchUri,
		Headers:   headers,
	}
}
