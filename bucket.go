package b2

import (
	"net/http"
	"time"
)

type Bucket struct {
	BucketID   string       `json:"bucketId"`
	BucketName string       `json:"bucketName"`
	BucketType BucketType   `json:"bucketType"`
	UploadUrls []*UploadUrl `json:"-"`
	B2         *B2          `json:"-"`
}

type BucketType string

const (
	AllPrivate BucketType = "allPrivate"
	AllPublic  BucketType = "allPublic"
)

// TODO? include some marker of being used (maybe mutex)
type UploadUrl struct {
	Url                string    `json:"uploadUrl"`
	AuthorizationToken string    `json:"authorizationToken"`
	Expiration         time.Time `json:"-"`
}

type listBucketsResponse struct {
	Buckets []Bucket `json:"buckets"`
}

type bucketRequest struct {
	AccountID  string     `json:"accountId"`
	BucketID   string     `json:"bucketId,omitempty"`
	BucketName string     `json:"bucketName,omitempty"`
	BucketType BucketType `json:"bucketType,omitempty"`
}

func (b2 *B2) ListBuckets() ([]Bucket, error) {
	req, err := b2.createBucketRequest("/b2api/v1/b2_list_buckets", bucketRequest{})
	if err != nil {
		return nil, err
	}

	resp, err := b2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b2.listBuckets(resp)
}

func (b2 *B2) listBuckets(resp *http.Response) ([]Bucket, error) {
	respBody := &listBucketsResponse{}
	err := ParseResponse(resp, respBody)
	if err != nil {
		return nil, err
	}

	for i := range respBody.Buckets {
		respBody.Buckets[i].B2 = b2
	}
	return respBody.Buckets, nil
}

func (b2 *B2) CreateBucket(name string, bucketType BucketType) (*Bucket, error) {
	bucketReq := bucketRequest{BucketName: name, BucketType: bucketType}
	req, err := b2.createBucketRequest("/b2api/v1/b2_list_buckets", bucketReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return b2.createBucket(resp)
}

func (b2 *B2) createBucket(resp *http.Response) (*Bucket, error) {
	bucket := &Bucket{B2: b2}
	err := ParseResponse(resp, bucket)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (b *Bucket) Update(newBucketType BucketType) error {
	br := bucketRequest{BucketID: b.BucketID, BucketType: newBucketType}
	req, err := b.B2.createBucketRequest("/b2api/v1/b2_update_bucket", br)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return b.update(resp)
}

func (b *Bucket) update(resp *http.Response) error {
	return ParseResponse(resp, b)
}

func (b *Bucket) Delete() error {
	br := bucketRequest{BucketID: b.BucketID}
	req, err := b.B2.createBucketRequest("/b2api/v1/b2_delete_bucket", br)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return b.bucketDelete(resp)
}

func (b *Bucket) bucketDelete(resp *http.Response) error {
	return ParseResponse(resp, b)
}

func (b2 *B2) createBucketRequest(path string, br bucketRequest) (*http.Request, error) {
	br.AccountID = b2.AccountID
	req, err := CreateRequest("POST", b2.ApiUrl+path, br)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b2.AuthorizationToken)
	return req, nil
}
