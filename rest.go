package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/schema"
	"go.uber.org/multierr"
)

type RESTClient interface {
	// Endpoints  add  endpoints to the client
	Endpoints(endpoint string) RESTClient

	Prefix(segments ...string) RESTClient

	Suffix(segments ...string) RESTClient

	Resource(resource string) RESTClient

	Name(resourceName string) RESTClient

	SubResource(subResources ...string) RESTClient

	Query(key string, value ...string) RESTClient

	Querys(value interface{}) RESTClient

	Headers(header http.Header) RESTClient

	Header(key string, values ...string) RESTClient

	Body(obj interface{}) RESTClient

	Retry(backoff backoff.BackOff,
		shouldRetryFunc func(*http.Response, error) bool) RESTClient

	Do(ctx context.Context, result interface{}, opts ...func(*http.Response) error) error

	DoNop(ctx context.Context, opts ...func(*http.Response) error) error

	DoRaw(ctx context.Context) ([]byte, error)

	Stream(ctx context.Context) (io.ReadCloser, error)
}

// NameMayNotBe specifies strings that cannot be used as names specified as path segments (like the REST API or etcd store)
var NameMayNotBe = []string{".", ".."}

// NameMayNotContain specifies substrings that cannot be used in names specified as path segments (like the REST API or etcd store)
var NameMayNotContain = []string{"/", "%"}

// IsValidPathSegmentName validates the name can be safely encoded as a path segment
func IsValidPathSegmentName(name string) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`may not be '%s'`, illegalName)}
		}
	}

	var errors []string
	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}

type restfulClient struct {
	c Transport

	From    func(context.Context) Logger
	baseURL *url.URL
	// generic components accessible via method setters
	verb         string
	pathPrefix   string
	subPath      string
	params       url.Values
	QueryEncoder *schema.Encoder
	headers      http.Header
	// retry
	backoff         backoff.BackOff
	shouldRetryFunc func(*http.Response, error) bool
	// structural elements of the request that are part of the Kubernetes API conventions
	resource     string
	resourceName string
	subresource  string

	// output
	err  error
	body io.Reader
}

// NewRESTClient start to reqest
func NewRESTClient(transport Transport, method string) *restfulClient {
	e := schema.NewEncoder()
	e.SetAliasTag("form")

	return &restfulClient{
		c:            transport,
		From:         Nop,
		verb:         method,
		QueryEncoder: e,
		backoff:      &backoff.ZeroBackOff{},
	}
}

func (r *restfulClient) AddError(err error) RESTClient {
	r.err = multierr.Append(r.err, err)
	return r
}

func (r *restfulClient) Endpoints(endpoint string) RESTClient {
	if endpoint == "" {
		return r
	}
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return r.AddError(err)
	}
	r.baseURL = baseURL
	return r
}

func (r *restfulClient) Prefix(segments ...string) RESTClient {
	r.pathPrefix = path.Join(r.pathPrefix, path.Join(segments...))
	return r
}

func (r *restfulClient) Suffix(segments ...string) RESTClient {
	r.subPath = path.Join(r.subPath, path.Join(segments...))
	return r
}

func (r *restfulClient) Resource(resource string) RESTClient {
	if len(r.resource) != 0 {
		return r.AddError(fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource))
	}
	if reasons := IsValidPathSegmentName(resource); len(reasons) != 0 {
		return r.AddError(fmt.Errorf("invalid resource %q: %v", resource, reasons))
	}
	r.resource = resource
	return r
}

func (r *restfulClient) Name(resourceName string) RESTClient {
	if len(resourceName) == 0 {
		return r.AddError(fmt.Errorf("resource name may not be empty"))
	}
	if len(r.resourceName) != 0 {
		return r.AddError(fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceName, resourceName))
	}
	if reasons := IsValidPathSegmentName(resourceName); len(reasons) != 0 {
		return r.AddError(fmt.Errorf("invalid resource name %q: %v", resourceName, reasons))
	}
	r.resourceName = resourceName
	return r
}

func (r *restfulClient) SubResource(subResources ...string) RESTClient {
	subresource := path.Join(subResources...)
	if len(r.subresource) != 0 {
		return r.AddError(fmt.Errorf("subresource already set to %q, cannot change to %q", r.subresource, subresource))
	}
	for _, s := range subResources {
		if reasons := IsValidPathSegmentName(s); len(reasons) != 0 {
			return r.AddError(fmt.Errorf("invalid subresource %q: %v", s, reasons))
		}
	}
	r.subresource = subresource
	return r
}

func (r *restfulClient) Query(key string, values ...string) RESTClient {
	if key == "" {
		return r
	}
	if r.params == nil {
		r.params = make(url.Values)
	}
	if len(values) == 0 {
		r.params.Del(key)
		return r
	}
	for _, value := range values {
		r.params.Add(key, value)
	}
	return r
}

func (r *restfulClient) Querys(value interface{}) RESTClient {
	if value == nil {
		return r
	}
	var form url.Values
	switch v := value.(type) {
	case url.Values:
		form = v
	case *url.Values:
		form = *v
	default:
		if err := r.QueryEncoder.Encode(value, form); err != nil {
			return r.AddError(err)
		}
	}
	if len(r.params) == 0 {
		r.params = form
		return r
	}
	r.params = make(url.Values, len(form))
	for key, srcValues := range form {
		dstValues, ok := r.params[key]
		if !ok {
			for _, srcValue := range srcValues {
				r.params.Add(key, srcValue)
			}
			continue
		}
		for _, srcValue := range srcValues {
			found := false
			for _, dstValue := range dstValues {
				if srcValue == dstValue {
					found = true
					break
				}
			}
			if !found {
				r.params.Add(key, srcValue)
			}
		}
	}
	return r
}

func (r *restfulClient) Headers(header http.Header) RESTClient {
	if len(r.headers) == 0 {
		r.headers = header
		return r
	}
	r.headers = make(http.Header, len(header))
	for key, srcValues := range header {
		dstValues, ok := r.headers[key]
		if !ok {
			for _, srcValue := range srcValues {
				r.headers.Add(key, srcValue)
			}
			continue
		}
		for _, srcValue := range srcValues {
			found := false
			for _, dstValue := range dstValues {
				if srcValue == dstValue {
					found = true
					break
				}
			}
			if !found {
				r.headers.Add(key, srcValue)
			}
		}
	}
	return r
}

func (r *restfulClient) Header(key string, values ...string) RESTClient {
	if key == "" {
		return r
	}
	if r.headers == nil {
		r.headers = http.Header{}
	}
	if len(values) == 0 {
		r.headers.Del(key)
		return r
	}
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

func (r *restfulClient) Body(obj interface{}) RESTClient {
	switch t := obj.(type) {
	case string:
		r.body = strings.NewReader(t)
	case []byte:
		r.body = bytes.NewReader(t)
	case io.Reader:
		r.body = t
	default:
		content, err := json.Marshal(t)
		if err != nil {
			return r.AddError(err)
		}
		r.body = bytes.NewReader(content)
		r.Header("Content-Type", "application/json; charset=utf-8")
	}
	return r
}

func (r *restfulClient) Retry(backoff backoff.BackOff,
	shouldRetryFunc func(*http.Response, error) bool) RESTClient {
	r.backoff = backoff
	r.shouldRetryFunc = shouldRetryFunc
	return r
}

func (r *restfulClient) roundTrip(req *http.Request, operate func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	body := req.Body
	defer body.Close()
	req.Body = io.NopCloser(body)
	var (
		err  error
		resp *http.Response
	)
	retryOperate := func() error {
		resp, err = operate(req)
		if r.shouldRetryFunc != nil && !r.shouldRetryFunc(resp, err) {
			return nil
		}
		return fmt.Errorf("attempt failed,%w", err)
	}

	backOff := backoff.WithContext(r.backoff, req.Context())

	notify := func(err error, duration time.Duration) {
		r.From(req.Context()).Warnf("attempt failed,duration %v, retry after %+v", duration, err)
	}

	backOff.Reset()
	if err := backoff.RetryNotify(retryOperate, backOff, notify); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *restfulClient) Do(ctx context.Context, result interface{}, opts ...func(*http.Response) error) error {
	if r.err != nil {
		return r.err
	}
	uri := r.finalURL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return err
	}
	var resp *http.Response
	if resp, err = r.roundTrip(req, r.c.RoundTrip); err != nil {
		return err
	}
	defer resp.Body.Close()
	return r.c.Response().Parse(resp, result, opts...)
}

func (r *restfulClient) DoNop(ctx context.Context, opts ...func(*http.Response) error) error {
	return r.Do(ctx, nil, opts...)
}

func (r *restfulClient) DoRaw(ctx context.Context) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	uri := r.finalURL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	if resp, err = r.roundTrip(req, r.c.RoundTrip); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (r *restfulClient) Stream(ctx context.Context) (io.ReadCloser, error) {
	if r.err != nil {
		return nil, r.err
	}
	uri := r.finalURL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	if resp, err = r.roundTrip(req, r.c.RoundTrip); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (r *restfulClient) finalURL() *url.URL {
	p := r.pathPrefix
	if len(r.resource) != 0 {
		p = path.Join(p, r.resource)
	}
	// Join trims trailing slashes, so preserve r.pathPrefix's trailing slash for backwards compatibility if nothing was changed
	if len(r.resourceName) != 0 || len(r.subPath) != 0 || len(r.subresource) != 0 {
		p = path.Join(p, r.resourceName, r.subresource, r.subPath)
	}
	finalURL := &url.URL{}
	if r.baseURL != nil {
		*finalURL = *r.baseURL
	}
	finalURL.Path = path.Join(finalURL.Path, p)
	query := finalURL.Query()
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}
