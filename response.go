package reqest

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

type Response interface {
	Parse(resp *http.Response, result interface{}, opts ...Func) error
}

type JsonResponse struct {
}

func (j JsonResponse) Parse(resp *http.Response, result interface{}, opts ...Func) error {
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if resp.StatusCode == http.StatusNoContent || result == nil {
		return nil
	}
	if err := j.checkContentType(resp.Header.Get("Content-Type")); err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	return decoder.Decode(result)
}

func (j JsonResponse) checkContentType(contentType string) error {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}
	if mediaType != "application/json" {
		return fmt.Errorf("can't parse content-type %s", contentType)
	}
	return nil
}
