package reqest

import (
	"net/http"
)

type Transport interface {
	http.RoundTripper
	WithRequest(requester Requester) Transport
	WithClient(roundTripper http.RoundTripper) Transport
	WithResponse(response Response) Transport

	Request() Requester
	Response() Response
	Client() http.RoundTripper

	Method(string) RESTClient
}

// DefaultTransport 默认配置的传输层实现
var DefaultTransport Transport = &transporter{
	req:          JsonRequest{},
	roundTripper: http.DefaultTransport,
	resp:         JsonResponse{},
}

type transporter struct {
	req          Requester
	roundTripper http.RoundTripper
	resp         Response
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
//
// RoundTrip should not attempt to interpret the response. In
// particular, RoundTrip must return err == nil if it obtained
// a response, regardless of the response's HTTP status code.
// A non-nil err should be reserved for failure to obtain a
// response. Similarly, RoundTrip should not attempt to
// handle higher-level protocol details such as redirects,
// authentication, or cookies.
//
// RoundTrip should not modify the request, except for
// consuming and closing the Request's Body. RoundTrip may
// read fields of the request in a separate goroutine. Callers
// should not mutate or reuse the request until the Response's
// Body has been closed.
//
// RoundTrip must always close the body, including on errors,
// but depending on the implementation may do so in a separate
// goroutine even after RoundTrip returns. This means that
// callers wanting to reuse the body for subsequent requests
// must arrange to wait for the Close call before doing so.
//
// The Request's URL and Header fields must be initialized.
func (t *transporter) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTripper.RoundTrip(req)
}

func (t *transporter) WithRequest(requester Requester) Transport {
	return &transporter{
		req:          requester,
		roundTripper: t.roundTripper,
		resp:         t.resp,
	}
}

func (t *transporter) WithClient(roundTripper http.RoundTripper) Transport {
	return &transporter{
		req:          t.req,
		roundTripper: roundTripper,
		resp:         t.resp,
	}
}

func (t *transporter) WithResponse(response Response) Transport {
	return &transporter{
		req:          t.req,
		roundTripper: t.roundTripper,
		resp:         response,
	}
}

func (t *transporter) Request() Requester {
	return t.req
}

func (t *transporter) Response() Response {
	return t.resp
}

func (t *transporter) Client() http.RoundTripper {
	return t.roundTripper
}

func (t *transporter) Method(method string) RESTClient {
	return &restfulClient{c: t, verb: method}
}

// Method start to reqest
func Method(method string) RESTClient {
	return &restfulClient{c: DefaultTransport, verb: method}
}
