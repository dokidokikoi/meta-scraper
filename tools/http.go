package tools

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// MakeRequest 创建一个远程请求对象
//
// 请求参数：
//
// method 字符串参数，传入请求类型，
// uri 字符串参数，传入请求地址，
// proxy 字符串参数，传入代理地址，
// body io读取接口，传入远程内容对象，
// header map结构，传入头部信息，
// cookies cookie数组，传入cookie信息。
//
// 返回数据：
//
// data 字节集，返回读取到的内容字节集，
// status 整数，返回请求状态码，
// err 错误信息。
func MakeRequest(
	method, uri, proxy string,
	body io.Reader,
	header map[string]string,
	cookies []*http.Cookie) (
	data []byte,
	status int,
	err error,
) {
	// 构建请求客户端
	client := createHTTPClient(proxy)

	// 创建请求对象
	req, err := createRequest(method, uri, body, header, cookies)
	// 检查错误
	if err != nil {
		return nil, 0, err
	}

	// 执行请求
	res, err := client.Do(req)
	// 检查错误
	if err != nil {
		return nil, 0, fmt.Errorf("%s [Request]: %s", uri, err)
	}

	// 获取请求状态码
	status = res.StatusCode
	// 读取请求内容
	data, err = ioutil.ReadAll(res.Body)
	// 关闭请求连接
	_ = res.Body.Close()

	return data, status, err
}

// 创建http客户端
func createHTTPClient(proxy string) *http.Client {
	// 初始化
	transport := &http.Transport{
		/* #nosec */
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 如果有代理
	if proxy != "" {
		// 解析代理地址
		proxyURI := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxy)
		}
		// 加入代理
		transport.Proxy = proxyURI
	}

	// 返回客户端
	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

// 创建请求对象
func createRequest(method, uri string, body io.Reader, header map[string]string, cookies []*http.Cookie) (*http.Request, error) {
	// 新建请求
	req, err := http.NewRequest(method, uri, body)
	// 检查错误
	if err != nil {
		return nil, fmt.Errorf("%s [Request]: %s", uri, err)
	}

	// 循环头部信息
	for k, v := range header {
		// 设置头部
		req.Header.Set(k, v)
	}

	// 设置了cookie
	if len(cookies) > 0 {
		// 循环cookie
		for _, cookie := range cookies {
			// 加入cookie
			req.AddCookie(cookie)
		}
	}

	return req, err
}
