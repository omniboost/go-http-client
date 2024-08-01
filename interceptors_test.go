package http_client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	http_client "github.com/omniboost/go-http-client"
)

func TestInterceptors(t *testing.T) {
	// Add interceptors
	client.Interceptors.AddRequestInterceptor(beforeRequestInterceptor)
	client.Interceptors.AddResponseInterceptor(onResponseInterceptor)

	// Create new request
	req, err := http.NewRequest(http.MethodGet, "https://profile.mcrhotels.com/api/properties", nil)
	if err != nil {
		t.Error(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "altxyTvvlNDEH7AoVy3NEr8A7SSdqP4YjFA6lffk"))

	var body interface{}
	resp, err := client.Do(req, body)
	if err != nil {
		t.Error(err)
	}

	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))
}

func beforeRequestInterceptor(client *http_client.Client, request *http.Request, body interface{}) error {
	fmt.Println("Performing request", request.URL.String())
	return nil
}

func onResponseInterceptor(client *http_client.Client, request *http.Request, response *http.Response) error {
	fmt.Println("After request", request.URL.String())
	return nil
}
