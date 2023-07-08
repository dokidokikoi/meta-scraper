package tools

import "net/url"

func AbsImage(base, uri string) string {
	baseURL, _ := url.Parse(base)
	absURL, _ := baseURL.Parse(uri)
	return absURL.String()
}
