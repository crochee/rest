package rest

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
)

// OnRetryCondition is a function to determine whether to retry
func OnRetryCondition(resp *http.Response, err error) bool {
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			// TCP连接断开或者传输有问题
			return true
		case errors.Is(err, context.Canceled):
			//  客户端主动取消
			return false
		default:
			var netErr net.Error
			if errors.As(err, &netErr) {
				// 发生网络上的错误
				return true
			}
			return false
		}
	}
	switch resp.StatusCode {
	case http.StatusBadGateway:
		return true
	case http.StatusGatewayTimeout:
		return true
	default:
	}
	return false
}
