package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"text/template"

	"github.com/pkg/errors"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-http-client/" + libraryVersion
)

// NewClient returns a new Exact Globe Client client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := &Client{}

	client.SetHTTPClient(httpClient)
	client.SetDebug(false)
	client.SetUserAgent(userAgent)

	return client
}

// Client manages communication
type Client struct {
	// HTTP client used to communicate with the Client.
	http *http.Client

	debug   bool
	baseURL url.URL

	// User agent for client
	userAgent string

	// Optional function interceptors after every request made to the DO Clients
	Interceptors ClientInterceptors
}

func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

func (c Client) Debug() bool {
	return c.debug
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c Client) BaseURL() url.URL {
	return c.baseURL
}

func (c *Client) SetBaseURL(baseURL url.URL) {
	c.baseURL = baseURL
}

func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

func (c Client) UserAgent() string {
	return userAgent
}

func (c *Client) GetEndpointURL(p string) url.URL {
	// Starting point is the Base URL
	clientURL := c.BaseURL()

	// Parse URL
	parsed, err := url.Parse(p)
	if err != nil {
		log.Fatal(err)
	}

	// Set the query
	q := clientURL.Query()
	for k, vv := range parsed.Query() {
		for _, v := range vv {
			q.Add(k, v)
		}
	}

	// Encode query
	clientURL.RawQuery = q.Encode()

	// Join Base URL with path
	clientURL.Path = path.Join(clientURL.Path, parsed.Path)
	return clientURL
}

func (c *Client) GetEndpointURLWithParams(p string, pathParams PathParams) url.URL {
	clientURL := c.GetEndpointURL(p)

	// Create path template
	tmpl, err := template.New("path").Parse(clientURL.Path)
	if err != nil {
		log.Fatal(err)
	}

	// Apply template
	buf := new(bytes.Buffer)
	params := pathParams.Params()
	err = tmpl.Execute(buf, params)
	if err != nil {
		log.Fatal(err)
	}

	// Apply path
	clientURL.Path = buf.String()
	return clientURL
}

func (c *Client) NewRequest(ctx context.Context, req Request) (*http.Request, error) {
	// convert body struct to json
	buf := new(bytes.Buffer)
	if req.RequestBodyInterface() != nil {
		err := json.NewEncoder(buf).Encode(req.RequestBodyInterface())
		if err != nil {
			return nil, err
		}
	}

	// create new http request
	r, err := c.NewRawRequest(ctx, req.Method(), req.URL().String(), buf)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) NewRawRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	// Create the new request
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Optionally pass along context
	if ctx != nil {
		r = r.WithContext(ctx)
	}

	// Set default headers
	r.Header.Add("User-Agent", c.UserAgent())

	return r, nil
}

// Do sends an Client request and returns the Client response. The Client response is json decoded and stored in the value
// pointed to by v, or returned as an error if an Client error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(req *http.Request, body interface{}) (*http.Response, error) {
	// Handle request interceptors
	err := c.Interceptors.handleRequestInterceptor(c, req, body)
	if err != nil {
		return nil, err
	}

	if c.debug {
		dump, _ := httputil.DumpRequestOut(req, true)
		log.Println(string(dump))
	}

	httpResp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle response interceptors
	err = c.Interceptors.handleResponseInterceptor(c, req, httpResp)
	if err != nil {
		return nil, err
	}

	// close body io.Reader
	defer func() {
		if rerr := httpResp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	if c.debug {
		dump, _ := httputil.DumpResponse(httpResp, true)
		log.Println(string(dump))
	}

	// check the provided interface parameter
	if httpResp == nil {
		return httpResp, nil
	}

	if body == nil {
		return httpResp, err
	}

	if httpResp.ContentLength == 0 {
		return httpResp, nil
	}

	status := &StatusResponse{Response: httpResp}
	// exResp := &ExceptionResponse{Response: httpResp}
	err = c.Unmarshal(httpResp.Body, []any{body}, []any{status})
	if err != nil {
		return httpResp, err
	}

	if status.Error() != "" {
		return httpResp, status
	}

	// check if the response isn't an error
	err = CheckResponse(httpResp)
	if err != nil {
		return httpResp, err
	}

	return httpResp, nil
}

func (c *Client) Unmarshal(r io.Reader, vv []interface{}, optionalVv []interface{}) error {
	if len(vv) == 0 && len(optionalVv) == 0 {
		return nil
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	for _, v := range vv {
		r := bytes.NewReader(b)
		dec := json.NewDecoder(r)

		err := dec.Decode(v)
		if err != nil && err != io.EOF {
			return errors.WithStack((err))
		}
	}

	for _, v := range optionalVv {
		r := bytes.NewReader(b)
		dec := json.NewDecoder(r)

		_ = dec.Decode(v)
	}

	return nil
}

// CheckResponse checks the Client response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range. Client error responses are expected to have either no response
// body, or a json response body that maps to ErrorResponse. Any other response
// body will be silently ignored.
func CheckResponse(r *http.Response) error {
	errorResponse := &ErrorResponse{Response: r}

	// Don't check content-lenght: a created response, for example, has no body
	// if r.Header.Get("Content-Length") == "0" {
	// 	errorResponse.Errors.Message = r.Status
	// 	return errorResponse
	// }

	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	// read data and copy it back
	data, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(data))
	if err != nil {
		return errorResponse
	}

	if r.ContentLength == 0 {
		return errors.New("response body is empty")
	}

	// convert json to struct
	if len(data) != 0 {
		err = json.Unmarshal(data, &errorResponse)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if errorResponse.Message != "" {
		return errorResponse
	}

	return nil
}

type StatusResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	Status  int    `json:"status"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
}

func (r *StatusResponse) Error() string {
	if r.Status != 0 && (r.Status < 200 || r.Status > 299) {
		if r.Msg == r.Message {
			return fmt.Sprintf("Status %d: %s", r.Status, r.Msg)
		}
		return fmt.Sprintf("Status %d: %s %s", r.Status, r.Msg, r.Message)
	}

	return ""
}

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	Message string `json:"Message"`
}

func (r *ErrorResponse) Error() string {
	return r.Message
}
