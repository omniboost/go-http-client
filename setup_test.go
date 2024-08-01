package http_client_test

import (
	"log"
	"net/url"
	"os"
	"testing"

	http_client "github.com/omniboost/go-http-client"
)

var (
	client *http_client.Client
)

func TestMain(m *testing.M) {
	baseURLString := os.Getenv("BASE_URL")
	debug := os.Getenv("DEBUG")

	client = http_client.NewClient(nil)
	if debug != "" {
		client.SetDebug(true)
	}

	if baseURLString != "" {
		baseURL, err := url.Parse(baseURLString)
		if err != nil {
			log.Fatal(err)
		}
		client.SetBaseURL(*baseURL)
	}
	m.Run()
}
