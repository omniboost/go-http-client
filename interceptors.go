package http_client

import "net/http"

// Request interceptor
type RequestInterceptor func(*Client, *http.Request, interface{}) error

// Response interceptor
type ResponseInterceptor func(*Client, *http.Request, *http.Response) error

type ClientInterceptors struct {
	request  []RequestInterceptor
	response []ResponseInterceptor
}

func (i *ClientInterceptors) AddRequestInterceptor(interceptor RequestInterceptor) {
	i.request = append(i.request, interceptor)
}

func (i *ClientInterceptors) AddResponseInterceptor(interceptor ResponseInterceptor) {
	i.response = append(i.response, interceptor)
}

func (i *ClientInterceptors) handleRequestInterceptor(client *Client, request *http.Request, body interface{}) error {
	for _, interceptor := range i.request {
		// Handle interceptor
		err := interceptor(client, request, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ClientInterceptors) handleResponseInterceptor(client *Client, request *http.Request, response *http.Response) error {
	for _, interceptor := range i.response {
		// Handle interceptor
		err := interceptor(client, request, response)
		if err != nil {
			return err
		}
	}

	return nil
}
