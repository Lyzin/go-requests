package nhr

import (
	"fmt"
	"strings"
)

// MontageUrl 拼接测试的接口URL，但是不拼接查询参数，主要是为了拼接最终的url
// 路径参数格式为列表 ["334","456"]
// 当路径参数没有时，拼接的路径为 https://host/apiUrl
// 当路径参数参数有时，按路径顺序拼接的路径为 https://host/apiUrl/pathParam/334/456
func MontageUrl(host, apiUrl string, pathParam ...interface{}) string {
	if string(apiUrl[0]) != "/" || strings.Contains(apiUrl, host) {
		return ""
	}

	if len(pathParam) == 0 {
		return fmt.Sprintf("https://%v%v", host, apiUrl)
	}
	var newPathParam string
	if len(pathParam) >= 1 {
		for _, v := range pathParam {
			newPathParam += fmt.Sprintf("/%v", v)
		}
	}
	return fmt.Sprintf("https://%v%v%v", host, apiUrl, newPathParam)
}
