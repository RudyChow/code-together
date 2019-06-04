package messages

import "encoding/json"

type Request Message

// GetRequest : 获取请求
func GetRequest(data []byte) (*Request, error) {
	var request *Request

	err := json.Unmarshal(data, &request)
	return request, err
}
