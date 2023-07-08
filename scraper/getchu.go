package scraper

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"scraper/tools"
	"strings"
	"time"
)

var (
	GetChuDomain = "https://www.getchu.com/"
)

type GetChu struct {
	Proxy     string
	Domain    string
	SearchUri string
	Headers   map[string]string
}

var GetChuScraper *GetChu

func (gc *GetChu) DoReq(url string) ([]byte, error) {
	data, status, err := tools.MakeRequest("GET", url, gc.Proxy, nil, gc.Headers, nil)
	if err != nil || status >= http.StatusBadRequest {
		fmt.Println("do http error status =", status)
		return nil, err
	}
	return data, nil
}

func (gc *GetChu) DoChromeReq(url string, headless bool, fs ...func(ctx context.Context)) ([]byte, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	chromeCtx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 保持浏览器窗口开启
	_ = chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)

	timeOutCtx, cancel := context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(timeOutCtx,
		network.Enable(),
		//需要爬取的网页的url
		chromedp.Navigate(url),
		network.SetExtraHTTPHeaders(map[string]interface{}{"Accept-Language": "zh-cn,zh;q=0.5", "X-Forwarded-For": "https://ggbases.dlgal.com/"}),
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	for _, f := range fs {
		f(timeOutCtx)
	}
	return []byte(htmlContent), err
}

func (gc *GetChu) GetItem(uri string) (*Item, error) {
	data, err := gc.DoReq(uri)
	if err != nil {
		return nil, err
	}
	item := &Item{Origin: uri}
	root, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// 获取名称
	item.Name, err = gc.GetItemName(root)
	if err != nil {
		fmt.Println("获取名称失败 url:", uri, "err:", err)
	}
	// 获取预览图
	item.Preview, err = gc.GetItemPreview(root)
	if err != nil {
		fmt.Println("获取预览图失败 url:", uri, "err:", err)
	}
	// 获取类别
	item.Genre, err = gc.GetItemGenre(root)
	if err != nil {
		fmt.Println("获取类别失败 url:", uri, "err:", err)
	}
	// 获取品牌
	item.Brand, err = gc.GetItemBrand(root)
	if err != nil {
		fmt.Println("获取品牌失败 url:", uri, "err:", err)
	}
	// 获取发售日
	item.ReleaseDate, err = gc.GetItemReleaseDate(root)
	if err != nil {
		fmt.Println("获取发售日失败 url:", uri, "err:", err)
	}
	// 获取官网链接
	item.Link, err = gc.GetItemLink(root)
	if err != nil {
		fmt.Println("获取官网链接失败 url:", uri, "err:", err)
	}
	// 获取故事简介链接
	item.Story, err = gc.GetItemStory(root)
	if err != nil {
		fmt.Println("获取故事简介失败 url:", uri, "err:", err)
	}
	// 获取角色信息
	item.Character, err = gc.GetItemCharacter(root)
	if err != nil {
		fmt.Println("获取角色信息失败 url:", uri, "err:", err)
	}

	return item, nil
}

func (gc *GetChu) GetItemName(node *goquery.Document) (string, error) {
	return strings.TrimSpace(tools.Jp2Utf8([]byte(node.Find("#soft-title").Text()))), nil
}

func (gc *GetChu) GetItemPreview(node *goquery.Document) ([]string, error) {
	var images []string

	node.Find("#soft_table a.highslide").Each(func(i int, selection *goquery.Selection) {
		if image, ok := selection.Attr("href"); ok {
			// 解析基本链接和相对链接
			images = append(images, tools.AbsImage(gc.Domain, image))
		}
	})

	node.Find("div.tabletitle").Each(func(i int, selection *goquery.Selection) {
		title := tools.Jp2Utf8([]byte(selection.Text()))
		if strings.Contains(title, "サンプル画像") {
			selection.Next().Find("a").Each(func(i int, a *goquery.Selection) {
				if image, ok := a.Attr("href"); ok {
					// 解析基本链接和相对链接
					images = append(images, tools.AbsImage(gc.Domain, image))
				}
			})
		}
	})
	return images, nil
}

func (gc *GetChu) GetItemGenre(node *goquery.Document) ([]string, error) {
	str := tools.Jp2Utf8([]byte(node.Find("#soft_table tr:nth-child(2) table tr:nth-child(5) td:nth-child(2)").
		Text()))
	return []string{str}, nil
}

func (gc *GetChu) GetItemBrand(node *goquery.Document) (string, error) {
	brand := node.Find("#soft_table tr:nth-child(2) table tr:nth-child(1) td:nth-child(2) a:nth-child(1)").
		Text()
	str := tools.Jp2Utf8([]byte(brand))
	return str, nil
}

func (gc *GetChu) GetItemReleaseDate(node *goquery.Document) (string, error) {
	return node.Find("#soft_table tr:nth-child(2) table tr:nth-child(3) td:nth-child(2) a").
		Text(), nil
}

func (gc *GetChu) GetItemLink(node *goquery.Document) (string, error) {
	link, ok := node.Find("#soft_table tr:nth-child(2) table tr:nth-child(1) td:nth-child(2) a:nth-child(1)").
		Attr("href")
	if !ok {
		return "", nil
	}
	return link, nil
}

func (gc *GetChu) GetItemStory(node *goquery.Document) (string, error) {
	var story string
	node.Find("div.tabletitle").Each(func(i int, selection *goquery.Selection) {
		title := tools.Jp2Utf8([]byte(selection.Text()))
		if strings.Contains(title, "ストーリー") {
			story = tools.Jp2Utf8([]byte(selection.Next().Text()))
			return
		}
	})
	return strings.TrimSpace(story), nil
}

func (gc *GetChu) GetItemCharacter(node *goquery.Document) ([]Character, error) {
	var character []Character
	node.Find("div.tabletitle").Each(func(i int, selection *goquery.Selection) {
		title := tools.Jp2Utf8([]byte(selection.Text()))
		if strings.Contains(title, "キャラクター") {
			trs := selection.Next().Find(`tr`)
			trs.Each(func(i int, selection *goquery.Selection) {
				if selection.Find("hr").Length() > 0 {
					return
				}
				avatar, _ := selection.Find("td:nth-child(1) img").Attr("src")
				name := tools.Jp2Utf8([]byte(selection.Find("td:nth-child(2) h2.chara-name").Text()))
				introduction := tools.Jp2Utf8([]byte(selection.Find("td:nth-child(2) dd").Text()))
				image, _ := selection.Find("td:nth-child(3) img").Attr("src")
				character = append(character, Character{
					Name:         name,
					Introduction: introduction,
					Avatar:       tools.AbsImage(gc.Domain, avatar),
					Images:       []string{tools.AbsImage(gc.Domain, image)},
				})
			})
			return
		}
	})

	return character, nil
}

func init() {
	headers := make(map[string]string)
	headers["User-Agent"] = defaultUserAgent
	headers["Referer"] = GetChuDomain
	headers["Accept-Language"] = "zh-CN,zh;q=0.9"
	headers["Cookie"] = "DLSESSIONID=e5135ab97299cbce3ef2dcf9bd188b69; _ga_RTF6MG3H5B=GS1.1.1687764148.2.1.1687764698.60.0.0; _gid=GA1.2.935234852.1688802045; getchu_adalt_flag=getchu.com; ITEM_HISTORY=1219845; _ga=GA1.1.58497094.1687400074; _ga_BSNR8334HV=GS1.1.1688802046.5.1.1688803436.57.0.0; _ga_JBMY6G3QFS=GS1.1.1688802046.1.1.1688803436.57.0.0"
	GetChuScraper = &GetChu{
		Proxy:     defaultProxy,
		Domain:    GetChuDomain,
		SearchUri: "",
		Headers:   headers,
	}
}
