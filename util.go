package b2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// Response Helpers
func ParseResponseBody(resp *http.Response, response interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return err
		}
		return errJson
	}

	return json.Unmarshal(body, response)
}
