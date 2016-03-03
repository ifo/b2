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
		replaceProtocol("https://api.backblaze.com/b2api/v1/b2_authorize_account"),
		nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(accountId, appKey)

	resp, err := clientDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		authJson := authResponse{}
		if err := json.Unmarshal(body, &authJson); err != nil {
			return nil, err
		}

		return &B2{
			AccountID:          authJson.AccountID,
			ApplicationKey:     appKey,
			AuthorizationToken: authJson.AuthorizationToken,
			ApiUrl:             authJson.ApiUrl,
			DownloadUrl:        authJson.DownloadUrl,
		}, nil
	} else {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return nil, err
		}

		return nil, errJson
	}
}
