package b2

import (
	"fmt"
	"net/http"
	"strings"
)

// basic httpClient and protocol - overwritten for test mocking
var httpClient = http.Client{}
var protocol = "https"

// global wrapper functions
func httpClientDo(req *http.Request) (*http.Response, error) {
	return httpClient.Do(req)
}

func replaceProtocol(url string) string {
	protoTrim := strings.Index(url, ":")
	if protoTrim == -1 {
		return url
	}
	return protocol + url[protoTrim:]
}

// TODO make error public, add checking for error type
type errorResponse struct {
	Status  int64  `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e errorResponse) Error() string {
	return fmt.Sprintf("Status: %d, Code: %s, Message: %s", e.Status, e.Code, e.Message)
}
