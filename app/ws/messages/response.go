package messages

type Response Message

func InfoResponse(data interface{}) *Response {
	response := &Response{
		Data: data,
	}
	response.Type = "info"
	return response
}

func RoomResponse(data interface{}) *Response {
	response := &Response{
		Data: data,
	}
	response.Type = "room"
	return response
}

func MessageResponse(data interface{}) *Response {
	response := &Response{
		Data: data,
	}
	response.Type = "message"
	return response
}

func CodeResponse(data interface{}) *Response {
	response := &Response{
		Data: data,
	}
	response.Type = "code"
	return response
}

func ChargeResponse(data interface{}) *Response {
	response := &Response{
		Data: data,
	}
	response.Type = "charge"
	return response
}
