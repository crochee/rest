# rest
restful api Go HTTP client,like k8s client-go
# Get Started
## install
You first need Go installed (version 1.17+ is required), then you can use the below Go command to install req:
```go
go get -u github.com/crochee/rest
```
## Usage
### Basic
```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/crochee/rest"
)

type Opts struct {
    Offset int `form:"offset"`
    Limit  int `form:"limit"`
}

func main() {
	var result struct {
		Content string `json:"content"`
	}
	if err := rest.Get().
        Endpoint("http://localhost:80").
        Prefix("v2").
        Resource("books").
        Query("limit", "20").
        Do(context.Background(), &result); err != nil {
            log.Println(err)
	}
    // curl 'http://localhost:80/v2/books?limit=20'

    if err := rest.Get().
    	Endpoint("http://localhost:80").
        Prefix("v2").
        Resource("books").
        Querys(Opts{Offset: 0, Limit: 20}).
        Do(context.Background(), &result); err != nil {
            log.Println(err)
	}
    // curl 'http://localhost:80/v2/books?limit=20'

    if err := rest.Post().
    	Endpoint("http://localhost:80").
        Prefix("v2").
        Resource("books").
        Body(map[string]interface{}{"name": "test"}]).
        Do(context.Background(), &result); err != nil {
            log.Println(err)
	}
    // curl -X POST 'http://localhost:80/v2/books' -d '{"name": "test"}'
}
```
### Multiple
client.go
```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/crochee/rest"
)

type IClient interface {
	Area() AreaSrv
}

func NewGateway() IClient {
	return &baseClient{reqest.NewHandler().Endpoint("http://localhost:80")}
}

type baseClient struct {
	rest.Handler
}

func (c baseClient) Area() AreaSrv {
	return areaSrv{c.Resource("areas")}
}

type AreaSrv interface {
	Get(ctx context.Context, id string) (*Areas, error)
}

type areaSrv struct {
	rest.Handler
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
If you have a bug report or feature request, you can [open an issue](https://github.com/crochee/rest/issues/new) or [pull request](https://github.com/crochee/rest/pulls).
