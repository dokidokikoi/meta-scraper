package tools

import (
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func Jp2Utf8(originBytes []byte) string {
	eucJPDecoder := japanese.EUCJP.NewDecoder()
	utf8Bytes, _, err := transform.Bytes(eucJPDecoder, originBytes)
	if err != nil {
		fmt.Println("转换失败:", err)
		return string(originBytes)
	}
	return string(utf8Bytes)
}
