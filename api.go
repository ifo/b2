package b2

import ()

type B2 struct {
	AccountID          string
	ApplicationKey     string
	AuthorizationToken string
	ApiUrl             string
	DownloadUrl        string
}

func MakeB2(accountId, appKey string) (*B2, error) {
	return &B2{}, nil
}
