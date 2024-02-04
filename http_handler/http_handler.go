package nhr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpRequests struct {
	Method   string
	URL      string
	Headers  map[string]string
	Cookies  []*http.Cookie
	Timeout  time.Duration
	PostBody string
	Params   string
}

type Option func(*HttpRequests)

// WithHeaders 设置请求头
func WithHeaders(headers map[string]string) Option {
	return func(req *HttpRequests) {
		req.Headers = headers
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(req *HttpRequests) {
		req.Timeout = timeout
	}
}

// WithCookies 设置cookies
func WithCookies(cookies []*http.Cookie) Option {
	return func(req *HttpRequests) {
		req.Cookies = cookies
	}
}

// WithParams 设置查询参数，且对url的参数进行encode
func WithParams(params map[string]string) Option {
	return func(req *HttpRequests) {
		// 将请求参数存入urlData中
		urlParams := url.Values{}
		for k, v := range params {
			urlParams.Set(k, v)
		}
		req.Params = urlParams.Encode() // URL encode
	}
}

// WithPostJsonBody
// 当headers的Content-Type是application/json
// HTTP会将请求参数以"键-值”"的方式组织的JSON格式数据，放到请求body里面
func WithPostJsonBody(data map[string]interface{}) Option {
	return func(req *HttpRequests) {
		dataToStr, err := json.Marshal(data)
		if err != nil {
			panic("convert postBody to string error")
			return
		}
		req.PostBody = string(dataToStr)
	}
}

// WithPostStringBody
// 当headers的Content-Type是application/x-www-form-urlencoded
// HTTP会将请求参数用key1=val1&key2=val2的方式进行组织，并放到请求body里面
func WithPostStringBody(data string) Option {
	return func(req *HttpRequests) {
		req.PostBody = data
	}
}

// createRequest 创建请求
func createRequest(requestIns *HttpRequests) *http.Response {
	// 将url转为URL结构体
	urlObj, err := url.ParseRequestURI(requestIns.URL)
	if err != nil {
		panic(fmt.Sprintf("parse url requestUrl failed, err:%v\n", err))
		return nil
	}
	// 将编码后的请求参数赋值给URL结构体的RawQuery字段
	// RequestObj.Params默认不传就是一个空字符串，要是用option模式传了，就走option模式来给Params字段赋值
	urlObj.RawQuery = requestIns.Params

	// 创建请求，这里需要注意：
	// 1、RequestObj.PostBody默认不传就是一个空字符串，要是用option模式传了，就走option模式来给PostBody字段赋值
	// 2、urlObj是URL结构体，并且它的查询请求参数已经被重新赋值过了，所以最终调用URL.String()方法就能拿到编码后的请求URL
	req, err := http.NewRequest(requestIns.Method, urlObj.String(), strings.NewReader(requestIns.PostBody))
	if err != nil {
		fmt.Println("create request instance failed")
		return nil
	}

	// 对上面创建的请求设置请求头
	// RequestObj.Headers不传就是默认的application/json
	// 要是用option模式传了，就走option模式来给Headers字段重新赋值
	for key, value := range requestIns.Headers {
		req.Header.Set(key, value)
	}

	// 添加登录的cookies
	if requestIns.Cookies != nil && len(requestIns.Cookies) > 0 {
		for _, v := range requestIns.Cookies {
			req.AddCookie(v)
		}
	}

	// 真正发起请求，返回http的response对象
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("send request error:%v\n", err.Error()))
		return nil
	}
	return response
}

// HttpCaller 发起请求
// method: HTTP method (GET, POST, PUT，DELETE)
// url: 请求的url
func HttpCaller(method, url string, options ...Option) *http.Response {
	RequestIns := &HttpRequests{
		// Method 请求方法转为大写
		Method: strings.ToUpper(method),

		// URL 请求URL
		URL: url,

		// 请求超时默认为3s
		Timeout: 3 * time.Second,

		// Headers的Content-Type默认为application/json
		Headers: map[string]string{"Content-Type": "application/json"},
	}

	// 通过option模式来设置HttpRequests的字段
	// 每一个opt都是func(*HttpRequests)类型，需要传入上面实例化的RequestObj，对RequestIns中的字段进行重新赋值
	for _, opt := range options {
		opt(RequestIns)
	}
	return createRequest(RequestIns)
}

// ResponseToBytes 将响应转为字节列表类型，可以反序列化为结构体
func responseToBytes(responseIns *http.Response) ([]byte, error) {
	defer responseIns.Body.Close()
	if responseIns.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status code not 200，actually status code is %v", responseIns.StatusCode)
	}
	bodyRet, err := ioutil.ReadAll(responseIns.Body)
	if err != nil {
		return nil, fmt.Errorf("read from response.Body failed:%v", err)
	}
	return bodyRet, nil
}

// ResponseToStruct 将字节切片类型的接口响应转接结构，通过结构体取值
// response：请求的响应对象
// v：结构体指针
func ResponseToStruct(responseIns *http.Response, v interface{}) error {
	responseBytesSlice, err := responseToBytes(responseIns)
	if err != nil {
		return fmt.Errorf("response to bytes error:%v", err)
	}
	err = FastJsonUnMarshal(responseBytesSlice, v)
	if err != nil {
		return fmt.Errorf("unMarshal response bytes slice error:%v", err)
	}
	return nil
}

// ResponseToMap 将响应结果转为map
func ResponseToMap(responseIns *http.Response) map[string]interface{} {
	defer responseIns.Body.Close()
	// 获取相应结果
	var ret map[string]interface{}
	err := json.NewDecoder(responseIns.Body).Decode(&ret)
	if err != nil {
		return nil
	}
	return ret
}
