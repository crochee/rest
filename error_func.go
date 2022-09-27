package reqest

import (
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

// DefaultFunc provides default error handling function implementation.
// If the implementation does not meet your needs, you can change it yourself
var DefaultFunc = []Func{ErrorFunc(http.StatusOK)}

func ErrorFunc(expectStatusCode int) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != expectStatusCode {
			decoder := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(resp.Body)
			decoder.UseNumber()
			var result struct {
				Code    string      `json:"code"`
				Message string      `json:"message"`
				Result  interface{} `json:"result"`
			}
			if err := decoder.Decode(&result); err != nil {
				return err
			}
			return fmt.Errorf("code:%s, message:%s, result:%v", result.Code, result.Message, result.Result)
		}
		return nil
	}
}
