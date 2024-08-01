package interceptors

import (
	"fmt"
	"net/http"
	"strings"

	http_client "github.com/omniboost/go-http-client"
)

func CheckContentType(contentType string) http_client.ResponseInterceptor {
	return func(client *http_client.Client, req *http.Request, resp *http.Response) error {
		header := resp.Header.Get("Content-Type")
		responseContentType := strings.Split(header, ";")[0]
		if responseContentType != contentType {
			return fmt.Errorf("expected Content-Type \"%s\", got \"%s\"", contentType, responseContentType)
		}

		return nil
	}
}
