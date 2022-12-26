package reqest

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

	"go.uber.org/multierr"
)

type RESTClient interface {
	Endpoints(endpoint string) RESTClient
	Prefix(segments ...string) RESTClient
	Suffix(segments ...string) RESTClient
	Resource(resource string) RESTClient
	Name(resourceName string) RESTClient
	SubResource(subResources ...string) RESTClient
	Query(key string, value ...string) RESTClient
	Header(key string, values ...string) RESTClient
	Body(obj interface{}) RESTClient
	Do(ctx context.Context, result interface{}, opts ...Func) error
	DoNop(ctx context.Context, opts ...Func) error
	DoReader(ctx context.Context) (io.Reader, error)
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

	baseURL *url.URL
	// generic components accessible via method setters
	verb       string
	pathPrefix string
	subPath    string
	params     url.Values
	headers    http.Header

	// structural elements of the request that are part of the Kubernetes API conventions
	resource     string
	resourceName string
	subresource  string

	// output
	err  error
	body io.Reader
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
	r.baseURL = baseURL
	return r.AddError(err)
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

func (r *restfulClient) Do(ctx context.Context, result interface{}, opts ...Func) error {
	if r.err != nil {
		return r.err
	}
	uri := r.URL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return err
	}
	var resp *http.Response
	if resp, err = r.c.RoundTrip(req); err != nil {
		return err
	}
	defer resp.Body.Close()
	return r.c.Response().Parse(resp, result, opts...)
}

func (r *restfulClient) DoNop(ctx context.Context, opts ...Func) error {
	return r.Do(ctx, nil, opts...)
}

func (r *restfulClient) DoReader(ctx context.Context) (io.Reader, error) {
	if r.err != nil {
		return nil, r.err
	}
	uri := r.URL().String()
	req, err := r.c.Request().Build(ctx, r.verb, uri, r.body, r.headers)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	if resp, err = r.c.RoundTrip(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, resp.Body); err != nil {
		return nil, err
	}
	return &buffer, nil
}

func (r *restfulClient) URL() *url.URL {
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
	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}
