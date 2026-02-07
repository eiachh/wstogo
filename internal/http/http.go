package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"
)

type Header map[string]string

type Request struct {
	Method  string
	URI     string
	Version string
	Headers Header
	Body    []byte
}

type Response struct {
	Version    string
	StatusCode int
	Status     string
	Headers    Header
	Body       []byte
}

func NewRequest(method, uri, version string, headers Header) *Request {
	if headers == nil {
		headers = make(Header)
	}
	return &Request{
		Method:  method,
		URI:     uri,
		Version: version,
		Headers: headers,
	}
}

func (r *Request) AddHeader(key, value string) {
	r.Headers[key] = value
}

func (r *Request) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, r.URI, r.Version))
	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	b.WriteString("\r\n")
	if len(r.Body) > 0 {
		b.Write(r.Body)
	}
	return b.String()
}

func ParseRequest(data []byte) (*Request, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	req := &Request{
		Method:  parts[0],
		URI:     parts[1],
		Version: parts[2],
		Headers: make(Header),
	}

	tp := textproto.NewReader(reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return nil, err
	}
	for k, v := range mimeHeader {
		if len(v) > 0 {
			req.Headers[k] = v[0]
		}
	}

	remaining, _ := io.ReadAll(reader)
	req.Body = remaining
	return req, nil
}

func ParseResponse(data []byte) (*Response, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid response line")
	}
	code := 0
	fmt.Sscanf(parts[1], "%d", &code)
	res := &Response{
		Version:    parts[0],
		StatusCode: code,
		Status:     strings.Join(parts[2:], " "),
		Headers:    make(Header),
	}

	tp := textproto.NewReader(reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return nil, err
	}
	for k, v := range mimeHeader {
		if len(v) > 0 {
			res.Headers[k] = v[0]
		}
	}

	remaining, _ := io.ReadAll(reader)
	res.Body = remaining
	return res, nil
}

func NewResponse(version string, statusCode int, status string, headers Header) *Response {
	if headers == nil {
		headers = make(Header)
	}
	return &Response{
		Version:    version,
		StatusCode: statusCode,
		Status:     status,
		Headers:    headers,
	}
}

func (r *Response) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%s %d %s\r\n", r.Version, r.StatusCode, r.Status))
	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	b.WriteString("\r\n")
	if len(r.Body) > 0 {
		b.Write(r.Body)
	}
	return b.String()
}

