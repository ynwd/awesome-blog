package res

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}, message string) Response {
	return Response{
		Status:  StatusSuccess,
		Message: message,
		Data:    data,
	}
}

func Error(message string) Response {
	return Response{
		Status:  StatusError,
		Message: message,
	}
}
