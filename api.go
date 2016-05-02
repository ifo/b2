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
	APIURL             string
	DownloadURL        string
	client             client
}

type client interface {
	Do(*http.Request) (*http.Response, error)
}

type authResponse struct {
	AccountID          string `json:"accountId"`
	AuthorizationToken string `json:"authorizationToken"`
	APIURL             string `json:"apiUrl"`
	DownloadURL        string `json:"downloadUrl"`
}

type ResponseError struct {
	Status  int64  `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Status: %d, Code: %s, Message: %s", e.Status, e.Code, e.Message)
}

func CreateB2(accountID, appKey string) (*B2, error) {
	b2 := &B2{
		AccountID:      accountID,
		ApplicationKey: appKey,
		client:         http.DefaultClient,
	}
	return b2.createB2()
}

func (b2 *B2) createB2() (*B2, error) {
	req, err := CreateRequest("GET", "https://api.backblaze.com/b2api/v1/b2_authorize_account", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(b2.AccountID, b2.ApplicationKey)
	resp, err := b2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b2.parseCreateB2Response(resp)
}

func (b2 *B2) parseCreateB2Response(resp *http.Response) (*B2, error) {
	ar := &authResponse{}
	err := parseResponse(resp, ar)
	if err != nil {
		return nil, err
	}
	b2.AuthorizationToken = ar.AuthorizationToken
	b2.APIURL = ar.APIURL
	b2.DownloadURL = ar.DownloadURL
	return b2, nil
}

func CreateRequest(method, url string, request interface{}) (*http.Request, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(method, url, bytes.NewReader(body))
}

func GetBzInfoHeaders(resp *http.Response) map[string]string {
	out := map[string]string{}
	for k, v := range resp.Header {
		if strings.HasPrefix(k, "X-Bz-Info-") {
			// strip Bz prefix and grab first header
			out[k[10:]] = v[0]
		}
	}
	return out
}

func parseResponse(resp *http.Response, body interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return parseResponseBody(resp, body)
	} else {
		return parseResponseError(resp)
	}
}

func parseResponseBody(resp *http.Response, body interface{}) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, body)
}

func parseResponseError(resp *http.Response) error {
	e := &ResponseError{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, e)
	if err != nil {
		return err
	}
	return e
}
