package reqest

import "net/http"

type Resource interface {
	Endpoint(endpoint string) Resource
	Resource(resource string) Resource
	To() Transport
}

func NewResource() Resource {
	return &ResourceHandler{}
}

type ResourceHandler struct {
	resource string
	endpoint string
}

func (r *ResourceHandler) Endpoint(endpoint string) Resource {
	return &ResourceHandler{
		resource: r.resource,
		endpoint: endpoint,
	}
}

func (r *ResourceHandler) Resource(resource string) Resource {
	return &ResourceHandler{
		resource: resource,
		endpoint: r.endpoint,
	}
}

func (r *ResourceHandler) To() Transport {
	return &fixTransport{
		t:        DefaultTransport,
		resource: r.resource,
		endpoint: r.endpoint,
	}
}

type fixTransport struct {
	t        Transport
	resource string
	endpoint string
}

func (t *fixTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t.t.RoundTrip(request)
}

func (t *fixTransport) WithRequest(requester Requester) Transport {
	return &fixTransport{
		t:        t.t.WithRequest(requester),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) WithClient(roundTripper http.RoundTripper) Transport {
	return &fixTransport{
		t:        t.t.WithClient(roundTripper),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) WithResponse(response Response) Transport {
	return &fixTransport{
		t:        t.t.WithResponse(response),
		resource: t.resource,
		endpoint: t.endpoint,
	}
}

func (t *fixTransport) Request() Requester {
	return t.t.Request()
}

func (t *fixTransport) Response() Response {
	return t.t.Response()
}

func (t *fixTransport) Method(method string) RESTClient {
	r := &restfulClient{c: t.t, verb: method}
	return r.Endpoints(t.endpoint).Resource(t.resource)
}

func (t *fixTransport) Get() RESTClient {
	return t.Method(http.MethodGet)
}

func (t *fixTransport) Post() RESTClient {
	return t.Method(http.MethodPost)
}

func (t *fixTransport) Delete() RESTClient {
	return t.Method(http.MethodDelete)
}

func (t *fixTransport) Put() RESTClient {
	return t.Method(http.MethodPut)
}

func (t *fixTransport) Patch() RESTClient {
	return t.Method(http.MethodPatch)
}
