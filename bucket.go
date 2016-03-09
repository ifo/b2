package b2

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Bucket struct {
	BucketID   string     `json:"bucketId"`
	BucketName string     `json:"bucketName"`
	BucketType BucketType `json:"bucketType"`
	B2         *B2        `json:"-"`
}

type BucketType string

const (
	AllPrivate BucketType = "allPrivate"
	AllPublic  BucketType = "allPublic"
)

type listBucketsResponse struct {
	Buckets []Bucket `json:"buckets"`
}

type bucketRequest struct {
	AccountID  string     `json:"accountId"`
	BucketID   string     `json:"bucketId,omitempty"`
	BucketName string     `json:"bucketName,omitempty"`
	BucketType BucketType `json:"bucketType,omitempty"`
}

func (b *B2) ListBuckets() ([]Bucket, error) {
	reqBody, err := json.Marshal(bucketRequest{AccountID: b.AccountID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST", b.MakeApiUrl("/b2api/v1/b2_list_buckets"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", b.AuthorizationToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		buckets := listBucketsResponse{}
		if err := json.Unmarshal(body, &buckets); err != nil {
			return nil, err
		}

		for i := range buckets.Buckets {
			buckets.Buckets[i].B2 = b
		}

		return buckets.Buckets, nil
	} else {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return nil, err
		}

		return nil, errJson
	}
}

func (b *B2) CreateBucket(name string, bType BucketType) (*Bucket, error) {
	reqBody, err := json.Marshal(bucketRequest{
		AccountID:  b.AccountID,
		BucketName: name,
		BucketType: bType,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST",
		b.MakeApiUrl("/b2api/v1/b2_create_bucket"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", b.AuthorizationToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		bucket := &Bucket{}
		if err := json.Unmarshal(body, bucket); err != nil {
			return nil, err
		}

		bucket.B2 = b

		return bucket, nil
	} else {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return nil, err
		}

		return nil, errJson
	}
}

func (b *Bucket) Update(newBucketType BucketType) error {
	reqBody, err := json.Marshal(bucketRequest{
		AccountID:  b.B2.AccountID,
		BucketID:   b.BucketID,
		BucketType: newBucketType,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		b.B2.MakeApiUrl("/b2api/v1/b2_update_bucket"), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", b.B2.AuthorizationToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		bucket := &Bucket{}
		if err := json.Unmarshal(body, bucket); err != nil {
			return err
		}

		b.BucketType = bucket.BucketType

		return nil
	} else {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return err
		}

		return errJson
	}
}

func (b *Bucket) Delete() error {
	reqBody, err := json.Marshal(bucketRequest{
		AccountID: b.B2.AccountID,
		BucketID:  b.BucketID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST",
		b.B2.MakeApiUrl("/b2api/v1/b2_delete_bucket"), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", b.B2.AuthorizationToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		bucket := &Bucket{}
		if err := json.Unmarshal(body, bucket); err != nil {
			return err
		}

		return nil
	} else {
		errJson := errorResponse{}
		if err := json.Unmarshal(body, &errJson); err != nil {
			return err
		}

		return errJson
	}
}
