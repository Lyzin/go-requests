package nhr

import jsoniter "github.com/json-iterator/go"

var fastJson = jsoniter.ConfigCompatibleWithStandardLibrary

// FastJsonMarshal json序列化
func FastJsonMarshal(v interface{}) ([]byte, error) {
	ret, err := fastJson.Marshal(v)
	return ret, err
}

// FastJsonUnMarshal json反序列化
func FastJsonUnMarshal(data []byte, v interface{}) error {
	err := fastJson.Unmarshal(data, v)
	return err
}
