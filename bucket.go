package b2

import (
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

type UploadUrl struct {
	Url                string    `json:"uploadUrl"`
	AuthorizationToken string    `json:"authorizationToken"`
	Time               time.Time `json:"-"`
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

func (b *B2) ListBuckets() ([]Bucket, error) {
	request := bucketRequest{AccountID: b.AccountID}
	response := &listBucketsResponse{}
	err := b.ApiRequest("POST", "/b2api/v1/b2_list_buckets", request, response)
	if err != nil {
		return nil, err
	}

	for i := range response.Buckets {
		response.Buckets[i].B2 = b
	}

	return response.Buckets, nil
}

func (b *B2) CreateBucket(name string, bType BucketType) (*Bucket, error) {
	request := bucketRequest{
		AccountID:  b.AccountID,
		BucketName: name,
		BucketType: bType,
	}
	response := &Bucket{B2: b}
	err := b.ApiRequest("POST", "/b2api/v1/b2_create_bucket", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (b *Bucket) Update(newBucketType BucketType) error {
	request := bucketRequest{
		AccountID:  b.B2.AccountID,
		BucketID:   b.BucketID,
		BucketType: newBucketType,
	}
	return b.B2.ApiRequest("POST", "/b2api/v1/b2_update_bucket", request, b)
}

func (b *Bucket) Delete() error {
	request := bucketRequest{
		AccountID: b.B2.AccountID,
		BucketID:  b.BucketID,
	}
	return b.B2.ApiRequest("POST", "/b2api/v1/b2_delete_bucket", request, &Bucket{})
}
