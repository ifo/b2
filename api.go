package b2

import (
	"encoding/json"
	"fmt"
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

type errorResponse struct {
	Status  int64  `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e errorResponse) Error() string {
	return fmt.Sprintf("Status: %d, Code: %s, Message: %s",
		e.Status, e.Code, e.Message)
}

func MakeB2(accountId, appKey string) (*B2, error) {
	req, err := http.NewRequest("GET",
		"https://api.backblaze.com/b2api/v1/b2_authorize_account", nil)
	if err != nil {
		return &B2{}, err
	}

	req.SetBasicAuth(accountId, appKey)

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return &B2{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &B2{}, err
	}

	if resp.StatusCode == 200 {
		authJson := authResponse{}
		if err := json.Unmarshal(body, authJson); err != nil {
			return &B2{}, err
		}

		return &B2{
			AccountID:          authJson.AccountID,
			ApplicationKey:     appKey,
			AuthorizationToken: authJson.AuthorizationToken,
			ApiUrl:             authJson.ApiUrl,
			DownloadUrl:        authJson.DownloadUrl,
		}, nil
	} else {
		errorJson := errorResponse{}
		if err := json.Unmarshal(body, errorJson); err != nil {
			return &B2{}, err
		}

		return &B2{}, errorJson
	}
}
