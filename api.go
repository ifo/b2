package b2

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type B2 struct {
	AccountID          string
	ApplicationKey     string
	AuthorizationToken string
	ApiUrl             string
	DownloadUrl        string
}

type authResponse struct {
	AccountID          string `json:"accountId"`
	AuthorizationToken string `json:"authorizationToken"`
	ApiUrl             string `json:"apiUrl"`
	DownloadUrl        string `json:"downloadUrl"`
}

func MakeB2(accountId, appKey string) (*B2, error) {
	b := &B2{
		AccountID:      accountId,
		ApplicationKey: appKey,
	}

	authResp := &authResponse{}
	req, err := b.CreateRequest("GET", "https://api.backblaze.com/b2api/v1/b2_authorize_account", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(accountId, appKey)
	err = b.DoRequest(req, authResp)
	if err != nil {
		return nil, err
	}

	b.AuthorizationToken = authResp.AuthorizationToken
	b.ApiUrl = authResp.ApiUrl
	b.DownloadUrl = authResp.DownloadUrl

	return b, nil
}

func (b *B2) ApiRequest(method, urlPart string, request, response interface{}) error {
	url := replaceProtocol(b.ApiUrl + urlPart)
	req, err := b.CreateRequest(method, url, request)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", b.AuthorizationToken)
	return b.DoRequest(req, response)
}

func (b *B2) CreateRequest(method, url string, request interface{}) (*http.Request, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, replaceProtocol(url), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (b *B2) DoRequest(req *http.Request, response interface{}) error {
	resp, err := httpClientDo(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
