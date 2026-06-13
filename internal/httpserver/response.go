package httpserver

type Response struct {
	Service string
	Status  string
	Content string
}

func NewResponse(status string, content string) *Response {
	return &Response{
		Service: "ssl_exporter",
		Status:  status,
		Content: content,
	}
}

var (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusOK      = "ok"
)
