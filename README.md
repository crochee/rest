# reqest
magic Go HTTP client,like k8s client-go
# Get Started
## install
You first need Go installed (version 1.17+ is required), then you can use the below Go command to install req:
```go
go get -u github.com/crochee/reqest
```
## Usage
### Basic
```go
package main

import (
	"context"
	"log"

	"github.com/crochee/reqest"
)

func main() {
	var result struct {
		Content string `json:"content"`
	}
	if err := reqest.
		DefaultTransport.
		Get().
		Prefix("v2").
		Query("page_size", "20").
		Do(context.Background(), &result); err != nil {
		log.Println(err)
	}
}
```
### Multiple
client.go
```go
package main

import (
	"context"
	"log"

	"github.com/crochee/reqest"
)

type IClient interface {
	Area() AreaSrv
}

func NewGateway() IClient {
	return &baseClient{reqest.NewHandler().Endpoint("http://localhost:80")}
}

type baseClient struct {
	reqest.Handler
}

func (c baseClient) Area() AreaSrv {
	return areaSrv{c.Resource("areas")}
}

type AreaSrv interface {
	Get(ctx context.Context, id string) (*Areas, error)
}

type areaSrv struct {
	reqest.Handler
}

func (a areaSrv) Get(ctx context.Context, id string) (*Areas, error) {
	var result Areas
	if err := a.To().
		Get().
		Prefix("v2").
		Name(id).
		Do(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type Areas struct {
	List     []string `json:"list"`
	PageNum  int      `json:"page_num"`
	PageSize int      `json:"page_size"`
	Total    int      `json:"total"`
}

func main() {
	result, err := NewGateway().Area().Get(context.Background(), "12")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
}
```
# Contributing
If you have a bug report or feature request, you can [open an issue](https://github.com/crochee/reqest/issues/new) or [pull request](https://github.com/crochee/reqest/pulls).
