package rest

import "net/http"

type Handler interface {
	Endpoint(endpoint string) Handler
	Resource(resource string) Handler
	To() Transport
}

func NewHandler() Handler {
	return &resourceHandler{}
}

type resourceHandler struct {
	resource string
	endpoint string
}

func (r *resourceHandler) Endpoint(endpoint string) Handler {
	return &resourceHandler{
		resource: r.resource,
		endpoint: endpoint,
	}
}

func (r *resourceHandler) Resource(resource string) Handler {
	return &resourceHandler{
		resource: resource,
		endpoint: r.endpoint,
	}
}

func (r *resourceHandler) To() Transport {
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

func (t *fixTransport) Client() http.RoundTripper {
	return t.t.Client()
}

func (t *fixTransport) Method(method string) RESTClient {
	return NewRESTClient(t.t, method).Endpoints(t.endpoint).Resource(t.resource)
}
