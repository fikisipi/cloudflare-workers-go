package structs

type Response interface {
	SetBody(body string) Response
	SetStatus(code int) Response
	AddHeader(key string, value string) Response
	Build() Response
}

type RawResponse struct {
	Body string
	StatusCode int
	Headers map[string]string
}

func (m *RawResponse) SetBody(body string) Response {
	m.Body = body; return m
}

func (m *RawResponse) SetStatus(code int) Response {
	m.StatusCode = code; return m
}

func (m *RawResponse) AddHeader(key string, value string) Response {
	m.Headers[key] = value; return m
}

func (m *RawResponse) Build() Response {
	return m
}