package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type B2 struct {
	AccountID          string
	ApplicationKey     string
	AuthorizationToken string
	ApiUrl             string
	DownloadUrl        string
	client             *client
}

type client struct {
	Protocol string
	http.Client
}

type authResponse struct {
	AccountID          string `json:"accountId"`
	AuthorizationToken string `json:"authorizationToken"`
	ApiUrl             string `json:"apiUrl"`
	DownloadUrl        string `json:"downloadUrl"`
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

func CreateB2(accountId, appKey string) (*B2, error) {
	c := &client{Protocol: "https", Client: http.Client{}}
	return createB2(accountId, appKey, c)
}

func createB2(accountId, appKey string, client *client) (*B2, error) {
	b := &B2{
		AccountID:      accountId,
		ApplicationKey: appKey,
		client:         client,
	}

	req, err := b.CreateRequest("GET", "https://api.backblaze.com/b2api/v1/b2_authorize_account", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(accountId, appKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	authResp := &authResponse{}
	err = ParseResponse(resp, authResp)
	if err != nil {
		return nil, err
	}

	b.AuthorizationToken = authResp.AuthorizationToken
	b.ApiUrl = authResp.ApiUrl
	b.DownloadUrl = authResp.DownloadUrl

	return b, nil
}

func (b *B2) CreateRequest(method, url string, request interface{}) (*http.Request, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	url, err = b.replaceProtocol(url)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (b *B2) replaceProtocol(url string) (string, error) {
	protoTrim := strings.Index(url, ":")
	if protoTrim == -1 {
		return url, fmt.Errorf("Invalid url")
	}
	return b.client.Protocol + url[protoTrim:], nil
}

func GetBzHeaders(resp *http.Response) map[string]string {
	out := map[string]string{}
	for k, v := range resp.Header {
		if strings.HasPrefix(k, "X-Bz-Info-") {
			// strip Bz prefix and grab first header
			out[k[10:]] = v[0]
		}
	}
	return out
}

func ParseResponse(resp *http.Response, respBody interface{}) error {
	if resp.StatusCode == 200 {
		return ParseResponseBody(resp, respBody)
	} else {
		return ParseErrorResponse(resp)
	}
}

func ParseResponseBody(resp *http.Response, respBody interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, respBody)
}

func ParseErrorResponse(resp *http.Response) error {
	errResp := &errorResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, errResp)
	if err != nil {
		return err
	}
	return errResp
}
