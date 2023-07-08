package scraper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"scraper/tools"
	"strconv"
	"sync"
	"time"
)

var GGBasesDomain = "https://ggbases.dlgal.com/"
var GGBasesSearchUri = "https://ggbases.dlgal.com/search.so?p=0&title=%d&advanced=0"
var GGBasesMagnetUri = "https://ggbases.dlgal.com/magnet.so?id=%s"
var GGBasesBtUri = "https://ggbases.dlgal.com/down.so?id=%s"

type GGBases struct {
	Proxy     string
	Domain    string
	SearchUri string
	Headers   map[string]string
}

var GGBasesScraper *GGBases

func (gg *GGBases) DoReq(url string) ([]byte, error) {
	data, status, err := tools.MakeRequest("GET", url, gg.Proxy, nil, gg.Headers, nil)
	if err != nil || status >= http.StatusBadRequest {
		fmt.Println("do http error status =", status)
		return nil, err
	}
	return data, nil
}

func (gg *GGBases) DoChromeReq(url string, headless bool, fs ...func(ctx context.Context)) ([]byte, error) {
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

func (gg *GGBases) GetItem(uri string) (*Item, error) {
	data, err := gg.DoChromeReq(uri, false)
	if err != nil {
		return nil, err
	}
	item := &Item{Origin: uri, proxy: gg.Proxy}
	root, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// 获取名称
	item.Name, err = gg.GetItemName(root)
	if err != nil {
		fmt.Println("获取名称失败 url:", uri, "err:", err)
	}
	// 获取预览图
	var errs []error
	item.Preview, errs = gg.GetItemPreviews(root)
	if errs != nil {
		fmt.Println("获取预览图失败 url:", uri, "err:", err)
	}
	// 获取标签
	item.Tags, err = gg.GetItemTags(root)
	if err != nil {
		fmt.Println("获取标签失败 url:", uri, "err:", err)
	}
	// 获取品牌
	item.Brand, err = gg.GetItemBrand(root)
	if err != nil {
		fmt.Println("获取品牌失败 url:", uri, "err:", err)
	}
	// 获取发售日期
	item.ReleaseDate, err = gg.GetItemReleaseDate(root)
	if err != nil {
		fmt.Println("获取发售日期失败 url:", uri, "err:", err)
	}
	// 获取官网链接
	item.Link, err = gg.GetItemBrand(root)
	if err != nil {
		fmt.Println("获取官网链接失败 url:", uri, "err:", err)
	}
	// 获取介绍页面
	item.Information, err = gg.GetItemInformation(root)
	if err != nil {
		fmt.Println("获取介绍页面失败 url:", uri, "err:", err)
	}
	// 获取存档
	item.SaveData, err = gg.GetItemSaveData(root)
	if err != nil {
		fmt.Println("获取存档失败 url:", uri, "err:", err)
	}
	// 获取攻略
	item.WalkThrough, err = gg.GetItemWalkThrough(root)
	if err != nil {
		fmt.Println("获取攻略失败 url:", uri, "err:", err)
	}
	// 获取大小
	item.Size, err = gg.GetItemSize(root)
	if err != nil {
		fmt.Println("获取大小失败 url:", uri, "err:", err)
	}
	// 获取磁链
	u, err := url.Parse(uri)
	if err == nil {
		params := u.Query()
		id := params.Get("id")

		item.Magnet, err = gg.GetItemMagnet(id)
		if err != nil {
			fmt.Println("获取磁链失败 url:", uri, "err:", err)
		}
	} else {
		fmt.Println("获取磁链失败 url:", uri, "err:", err)
	}
	// 获取 bt 文件
	u, err = url.Parse(uri)
	if err == nil {
		params := u.Query()
		id := params.Get("id")

		item.BtFile, err = gg.GetItemBtFile(id)
		if err != nil {
			fmt.Println("获取bt文件失败 url:", uri, "err:", err)
		}
	} else {
		fmt.Println("获取bt文件失败 url:", uri, "err:", err)
	}
	// 获取其他信息
	item.OtherInfo, err = gg.GetItemOtherInfo(root)
	if err != nil {
		fmt.Println("获取其他信息失败 url:", uri, "err:", err)
	}
	return item, nil
}

func (gg GGBases) GetItemName(node *goquery.Document) (string, error) {
	return node.Find("#atitle").Text(), nil
}

func (gg GGBases) GetItemPreviews(node *goquery.Document) ([]string, []error) {
	var errs []error
	td := node.Find("#touch tbody>tr:nth-child(7)>td")
	linkStart, ok := td.Find("#showCoverBtn").Attr("href")
	if !ok {
		return nil, []error{errors.New("获取封面图片失败")}
	}
	numStr := td.Find("#showCoverBtn span").Text()
	re := regexp.MustCompile(`(\d+)`)              // 匹配括号中的数字，并用 () 分组
	matches := re.FindAllStringSubmatch(numStr, 1) // 获取匹配的分组
	if len(matches) <= 0 {
		return nil, []error{errors.New("无预览图")}
	}
	num, err := strconv.Atoi(matches[0][1])
	if err != nil {
		return nil, append(errs, err)
	}
	images := make([]string, 0, num)
	wait := sync.WaitGroup{}
	var lock sync.Mutex
	for i := 0; i < num; i++ {
		wait.Add(1)
		go func(i int) {
			image, err := gg.GetItemImage("https:" + linkStart + "/" + strconv.Itoa(i))
			if err != nil {
				errs = append(errs, err)
				return
			}
			lock.Lock()
			images = append(images, image)
			lock.Unlock()
			wait.Done()
		}(i)
	}
	wait.Wait()
	return images, errs
}

func (gg *GGBases) GetItemImage(url string) (string, error) {
	data, err := gg.DoChromeReq(url, true)
	if err != nil {
		return "", err
	}
	node, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	image, _ := node.Find("#showpictd img").Attr("src")
	return fmt.Sprintf("https:%s", image), nil
}

func (gg GGBases) GetItemTags(node *goquery.Document) ([]Tag, error) {
	var tags []Tag
	node.Find("#extagstable tbody tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}
		tag := Tag{
			Category: Category{},
			Item:     []TagItem{},
		}
		tag.Category.Name = tr.Find("td:nth-child(1)>a").Text()

		tr.Find("td:nth-child(2) a").Each(func(i int, a *goquery.Selection) {
			var ok bool
			tagItem := TagItem{}
			tagItem.Name = a.Find("span").Text()
			tagItem.Identity, ok = a.Attr("title")
			if !ok {
				tagItem.Identity = tagItem.Name
			}
			tag.Item = append(tag.Item, tagItem)
		})
		tags = append(tags, tag)
	})

	return tags, nil
}

func (gg GGBases) GetItemBrand(node *goquery.Document) (string, error) {
	return "", nil
}

func (gg GGBases) GetItemReleaseDate(node *goquery.Document) (string, error) {
	return node.Find("#touch tbody tr:nth-child(5) td:nth-child(1) span").Text(), nil
}
func (gg GGBases) GetItemLink(node *goquery.Document) (string, error) {
	return "", nil
}
func (gg GGBases) GetItemInformation(node *goquery.Document) ([]string, error) {
	return nil, nil
}
func (gg GGBases) GetItemSaveData(node *goquery.Document) (string, error) {
	link, _ := node.Find("#touch tbody>tr:nth-child(7)>td a:nth-child(2)").Attr("href")
	return fmt.Sprintf("https:%s", link), nil
}
func (gg GGBases) GetItemWalkThrough(node *goquery.Document) (string, error) {
	link, _ := node.Find("#touch tbody>tr:nth-child(7)>td a:nth-child(3)").Attr("href")
	return fmt.Sprintf("https:%s", link), nil
}
func (gg GGBases) GetItemSize(node *goquery.Document) (string, error) {
	return node.Find("#touch tbody tr:nth-child(5) td:nth-child(2) span").Text(), nil
}
func (gg GGBases) GetItemMagnet(id string) (string, error) {
	data := make(chan []byte, 1)
	f := func(ctx context.Context) {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch ev := ev.(type) {

			case *network.EventResponseReceived:
				if ev.Type != "XHR" {
					return
				}
				resp := ev.Response
				if resp.Status > http.StatusBadRequest {
					fmt.Println(resp.Status)
					return
				}
				fmt.Println(resp.Status)
				go func() {
					c := chromedp.FromContext(ctx)
					buf, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
					if err != nil {
						fmt.Println("get xhr resp body err:", err)
					}
					data <- buf
				}()
			}
		})
		chromedp.Run(ctx,
			network.Enable(),
			chromedp.Click(".dbutton[bt='3']"),
			chromedp.Sleep(time.Second*1),
		)
	}
	_, err := gg.DoChromeReq(fmt.Sprintf(GGBasesMagnetUri, id), false, f)
	if err != nil {
		return "", err
	}
	body := <-data
	fmt.Println("magnet =", string(body))
	hash := gjson.GetBytes(body, "hash").String()
	return fmt.Sprintf("magnet:?xt=urn:btih:%s", hash), nil
}

func (gg GGBases) GetItemBtFile(id string) (string, error) {
	done := make(chan string, 1)
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	f := func(ctx context.Context) {
		chromedp.ListenTarget(ctx, func(v interface{}) {
			if ev, ok := v.(*browser.EventDownloadProgress); ok {
				if ev.State == browser.DownloadProgressStateCompleted {
					done <- ev.GUID
					close(done)
				}
			}
		})
		chromedp.Run(ctx,
			network.Enable(),
			browser.
				SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
				WithDownloadPath(wd).
				WithEventsEnabled(true),
			chromedp.Click(".dbutton[bt='1']"),
			chromedp.Sleep(time.Second),
		)
	}
	_, err = gg.DoChromeReq(fmt.Sprintf(GGBasesBtUri, id), false, f)
	return "", err
}
func (gg GGBases) GetItemOtherInfo(node *goquery.Document) (string, error) {
	return node.Find("#description div[markdown-text]").Html()
}

func (gg *GGBases) SetHeader(k, v string) {
	gg.Headers[k] = v
}

func init() {
	headers := make(map[string]string)
	headers["User-Agent"] = defaultUserAgent
	headers["Referer"] = GGBasesDomain
	headers["Accept-Language"] = "zh-CN,zh;q=0.9"
	GGBasesScraper = &GGBases{
		Proxy:     defaultProxy,
		Domain:    GGBasesDomain,
		SearchUri: GGBasesSearchUri,
		Headers:   headers,
	}
}
