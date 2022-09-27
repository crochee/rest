package reqest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Requester interface {
	Build(ctx context.Context, method, url string, body interface{}, headers http.Header) (*http.Request, error)
}

type JsonRequest struct {
}

func (j JsonRequest) Build(ctx context.Context, method string, url string, body interface{}, headers http.Header) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		switch data := body.(type) {
		case string:
			reader = strings.NewReader(data)
		case []byte:
			reader = bytes.NewReader(data)
		case io.Reader:
			reader = data
		default:
			content, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewReader(content)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	return req, nil
}
