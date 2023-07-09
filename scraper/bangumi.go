package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"scraper/tools"
)

var (
	bangumiToken     = ""
	bangumiUserAgent = "dokidokikoi/meta-scraper (https://github.com/dokidokikoi/meta-scraper)"

	BangumiDomain = "https://api.bgm.tv/"
)

type Bangumi struct {
	Proxy     string
	Domain    string
	SearchUri string
	Headers   map[string]string
}

var BangumiScraper *Bangumi

func (b *Bangumi) DoReq(method, uri string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	data, _, err = tools.MakeRequest(method, uri, b.Proxy, bytes.NewBuffer(data), b.Headers, nil)
	return data, err
}

//		proxy       string
//		Name        string      // 名称
//		Cover       string      // 封面
//		Preview     []string    // 预览图
//		Tags        []Tag       // 标签
//		Brand       string      // 品牌
//		ReleaseDate string      // 发售日
//		Link        string      // 官网
//		Information []string    // 介绍页面
//		SaveData    string      // 存档
//		WalkThrough string      // 攻略
//		Size        string      // 大小（仅供参考）
//		Magnet      string      // 磁力链接
//		BtFile      string      // bt 种子
//		OtherInfo   string      // 其它信息
//		Origin      string      // 来源网站
//		Character   []Character // 角色
//		Genre       []string    // 类别
//		Story       string      // 故事简介
//	}
func (b *Bangumi) GetItem(uri string) (*Item, error) {
	data, err := b.DoReq("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	item := &Item{Origin: uri}
	// 获取名称
	item.Name, err = b.GetItemName(data)
	if err != nil {
		fmt.Println("获取名称失败 url:", uri, "err:", err)
	}
	// 获取预览图
	item.Preview, err = b.GetItemPreview(data)
	if err != nil {
		fmt.Println("获取预览图失败 url:", uri, "err:", err)
	}
	// 获取类别
	item.Genre, err = b.GetItemGenre(data)
	if err != nil {
		fmt.Println("获取类别失败 url:", uri, "err:", err)
	}
	// 获取品牌
	item.Brand, err = b.GetItemBrand(data)
	if err != nil {
		fmt.Println("获取品牌失败 url:", uri, "err:", err)
	}
	// 获取发售日
	item.ReleaseDate, err = b.GetItemReleaseDate(data)
	if err != nil {
		fmt.Println("获取发售日失败 url:", uri, "err:", err)
	}
	// 获取官网链接
	item.Link, err = b.GetItemLink(data)
	if err != nil {
		fmt.Println("获取官网链接失败 url:", uri, "err:", err)
	}
	// 获取故事简介链接
	item.Story, err = b.GetItemStory(data)
	if err != nil {
		fmt.Println("获取故事简介失败 url:", uri, "err:", err)
	}
	// 获取角色信息
	var errs []error
	id := gjson.GetBytes(data, "id").String()
	item.Character, errs = b.GetItemCharacter(id)
	if errs != nil {
		fmt.Println("获取角色信息失败 url:", uri, "err:", errs)
	}
	return item, nil
}

func (b *Bangumi) GetItemName(data []byte) (string, error) {
	return gjson.GetBytes(data, "name").String(), nil
}

func (b *Bangumi) GetItemPreview(data []byte) ([]string, error) {
	return []string{gjson.GetBytes(data, "images.large").String()}, nil
}

func (b *Bangumi) GetItemGenre(data []byte) ([]string, error) {
	for _, info := range gjson.GetBytes(data, "infobox").Array() {
		if info.Get("key").String() == "游戏类型" {
			return []string{info.Get("value").String()}, nil
		}
	}
	return nil, errors.New("未匹配游戏类型")
}

func (b *Bangumi) GetItemBrand(data []byte) (string, error) {
	for _, info := range gjson.GetBytes(data, "infobox").Array() {
		if info.Get("key").String() == "开发" {
			return info.Get("value").String(), nil
		}
	}
	return "", errors.New("未匹配游戏品牌")
}

func (b *Bangumi) GetItemReleaseDate(data []byte) (string, error) {
	for _, info := range gjson.GetBytes(data, "infobox").Array() {
		if info.Get("key").String() == "发行日期" {
			return info.Get("value").String(), nil
		}
	}
	return "", errors.New("未匹配游戏发行日期")
}

func (b *Bangumi) GetItemLink(data []byte) (string, error) {
	for _, info := range gjson.GetBytes(data, "infobox").Array() {
		if info.Get("key").String() == "website" {
			return info.Get("value").String(), nil
		}
	}
	return "", errors.New("未匹配游戏官网链接")
}

func (b *Bangumi) GetItemStory(data []byte) (string, error) {
	return gjson.GetBytes(data, "summary").String(), nil
}

func (b Bangumi) GetItemCharacter(id string) ([]Character, []error) {
	var errs []error
	var characters []Character
	data, err := b.DoReq("GET", fmt.Sprintf("https://api.bgm.tv/v0/subjects/%s/characters", id), nil)
	if err != nil {
		return nil, append(errs, fmt.Errorf("发送获取角色请求失败 err %v", err))
	}

	for _, c := range gjson.ParseBytes(data).Array() {
		cid := c.Get("id").String()
		if cid != "" {
			data, err = b.DoReq("GET", fmt.Sprintf("https://api.bgm.tv/v0/characters/%s", cid), nil)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			characters = append(characters, Character{
				Name:         gjson.GetBytes(data, "name").String(),
				Introduction: gjson.GetBytes(data, "summary").String(),
				Avatar:       gjson.GetBytes(data, "images.large").String(),
			})
		}
	}
	return characters, errs
}

func init() {
	headers := make(map[string]string)
	headers["User-Agent"] = bangumiUserAgent
	headers["Authorization"] = fmt.Sprintf("Bearer %s", bangumiToken)
	BangumiScraper = &Bangumi{
		Proxy:     defaultProxy,
		Domain:    BangumiDomain,
		SearchUri: "",
		Headers:   headers,
	}
}
