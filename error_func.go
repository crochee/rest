package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ErrorFunc(expectStatusCode int) func(*http.Response) error {
	return func(resp *http.Response) error {
		if resp.StatusCode != expectStatusCode {
			decoder := json.NewDecoder(resp.Body)
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
