package scraper

//func main() {
//	headers := make(map[string]string)
//	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
//	headers["Referer"] = "https://nhentai.net/g/462159/"
//	headers["Cookie"] = "cf_clearance=HxuE9VwkVZnvkoDGLYrKWxe_qKv0ORD3SJVlLoFXJ1Q-1688575031-0-160;"
//	data, status, err := tools.MakeRequest("GET", "https://nhentai.net/g/462159/", "socks5://127.0.0.1:7890", nil, headers, nil)
//	if err != nil {
//		fmt.Println("do http error", err)
//		return
//	}
//	fmt.Println("status=", status)
//	// 转换为节点数据
//	root, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
//	if err != nil {
//		fmt.Println("parse data error", err)
//	}
//	fmt.Println(root.Html())
//}
