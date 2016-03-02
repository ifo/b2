package b2

import (
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

	// TODO handle response errors

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &B2{}, err
	}

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
}
