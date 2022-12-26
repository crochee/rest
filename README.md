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
import (
    "context"
    
    "github.com/crochee/reqest"
)

func foo(ctx context.Context) {
	var result struct {
		Content string `json:"content"`
	}
	return reqest.DefaultTransport.
		Method(http.MethodGet).
		Prefix("v2").
		Query("flavor_types").
		Query("sys_volume_types", "SysVolumeTypes").
		Do(context.Background(), &result)
}

```
### Multiple
client.go
```go
import (
    "github.com/crochee/reqest"
)

type IClient interface {
    Area() AreaSrv
}

var iClient IClient

func SetClient(c IClient) {
    iClient = c
}

func NewClient() IClient {
    return iClient
}


func NewBaseClient() IClient {
    return &baseClient{reqest.NewResource().Endpoint("http://localhost:80")}
}

type baseClient struct {
    reqest.Resource
}

func (c baseClient) Area() AreaSrv {
    return Area{c.Resource("areas")}
}
```
area.go
```go
import (
    "log"
	
    "github.com/crochee/reqest"
)

type AreaSrv interface {
    List(ctx context.Context) error
}

type Area struct {
    reqest.Resource
}

func (a Area) List(ctx context.Context) error {
    var result interface{}
    if err := a.To().
        Get().
        Prefix("v2").
        Query("limit", "20").
        Query("offset", "0").
        Do(ctx, &result); err != nil {
        return err
    }
    log.Println(result)
    return nil
}
```
# Contributing
If you have a bug report or feature request, you can [open an issue](https://github.com/crochee/reqest/issues/new) or [pull request](https://github.com/crochee/reqest/pulls).
