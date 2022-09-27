package reqest

import (
	"fmt"
	"mime"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Func func(*http.Response) error

type Response interface {
	Parse(resp *http.Response, result interface{}, opts ...Func) error
}

type JsonResponse struct {
}

func (j JsonResponse) Parse(resp *http.Response, result interface{}, opts ...Func) error {
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if len(opts) == 0 {
		opts = append(opts, DefaultFunc...)
	}
	for _, opt := range opts {
		if err := opt(resp); err != nil {
			return err
		}
	}
	if result == nil {
		return nil
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	if mediaType != "application/json" {
		return fmt.Errorf("can't parse body with %s", mediaType)
	}
	decoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(resp.Body)
	decoder.UseNumber()
	return decoder.Decode(result)
}
