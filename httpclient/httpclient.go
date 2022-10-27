// Package httpclient provides a simple http client interface to access web resource.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client is a simple wrap of http client.
// It provides a simple way to handle cookies, UA, and timeouts.
type Client struct {
	client http.Client

	Timeout time.Duration
	UA      string
}

// A Request represents an HTTP request received by a server
// or to be sent by a client.
// It contains a vlaid request if Err != nil.
type Request struct {
	*http.Request
	Err    error
	Cancel func()
	Client *Client
}

// Response represents the response from an HTTP request.
// It contains a valid http response if Err != nil.
type Response struct {
	*http.Response
	Request *Request
	Err     error
}

// Invode sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client.
func (req *Request) Invoke() *Response {
	if req.Err != nil {
		return &Response{Request: req, Err: req.Err}
	}
	rsp, err := req.Client.client.Do(req.Request)
	if err != nil {
		return &Response{Request: req, Err: err}
	}
	return &Response{Response: rsp, Request: req}
}

// Drop drop the response body and return a nil.
// It also returns error if HTTP response status code is not 200.
func (rsp *Response) Drop() error {
	if rsp.Err != nil {
		return rsp.Err
	}
	defer rsp.Request.Cancel()
	defer rsp.Body.Close()
	return rsp.checkStatusCode()
}

// JSON unmarshal HTTP response body into input object.
// It also returns error if HTTP response status code is not 200.
func (rsp *Response) JSON(o interface{}) error {
	if rsp.Err != nil {
		return rsp.Err
	}
	defer rsp.Request.Cancel()
	defer rsp.Body.Close()
	if err := rsp.checkStatusCode(); err != nil {
		return err
	}
	dec := json.NewDecoder(rsp.Body)
	return dec.Decode(o)
}

// WriteTo writes HTTP response body into a writer.
// It also returns error if HTTP response status code is not 200.
func (rsp *Response) WriteTo(w io.Writer) (int64, error) {
	if rsp.Err != nil {
		return 0, rsp.Err
	}
	defer rsp.Request.Cancel()
	defer rsp.Body.Close()
	if err := rsp.checkStatusCode(); err != nil {
		return 0, err
	}
	return io.Copy(w, rsp.Body)
}

func (rsp *Response) checkStatusCode() error {
	if rsp.StatusCode != 200 {
		data, _ := io.ReadAll(rsp.Body)
		return fmt.Errorf("Response %d, body=%s", rsp.StatusCode, string(data))
	}
	return nil
}

// Options for creating a Client.
type Options struct {
	// Proxy sets a http proxy. It uses environments if is empty.
	// The proxy type is determined by the URL scheme. "http",
	// "https", and "socks5" are supported. If the scheme is empty,
	// "http" is assumed.
	Proxy string
	// UA of HTTP request header User-Agent.
	UA string
	// Timeout of HTTP request.
	Timeout time.Duration
}

// New creates a new Client with given option.
func New(o *Options) (*Client, error) {
	var c Client
	jar, _ := cookiejar.New(nil)
	c.client.Jar = jar
	c.Timeout = time.Second * 30
	c.UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36"
	if o != nil {
		if o.Proxy != "" {
			u, err := url.Parse(o.Proxy)
			if err != nil {
				return nil, err
			}
			c.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(u),
			}
		}
		if o.Timeout > 0 {
			c.Timeout = o.Timeout

		}
		if o.UA != "" {
			c.UA = o.UA
		}
	}
	return &c, nil
}

func (c *Client) setDefault(r *http.Request) {
	r.Header.Set("User-Agent", c.UA)
}

// Get creates HTTP GET request.
func (c *Client) Get(url string) *Request {
	return c.makeRequest("GET", url, nil)
}

func (c *Client) makeRequest(method, url string, body io.Reader) *Request {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	req, err := http.NewRequestWithContext(ctx, "GET", url, body)
	if err != nil {
		cancel()
		return &Request{Err: err}
	}
	c.setDefault(req)
	return &Request{Request: req, Client: c, Cancel: cancel}
}

// PostJSON creates a HTTP POST request with json body.
func (c *Client) PostJSON(url string, body interface{}) *Request {
	data, err := json.Marshal(body)
	if err != nil {
		return &Request{Err: err}
	}
	req := c.makeRequest("POST", url, bytes.NewReader(data))
	if req.Err != nil {
		return req
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// PostForm creates a HTTP POST request with form body.
func (c *Client) PostForm(url string, values url.Values) *Request {
	req := c.makeRequest("POST", url, strings.NewReader(values.Encode()))
	if req.Err != nil {
		return req
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
