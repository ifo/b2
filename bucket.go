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

func (b *B2) ListBuckets() ([]Bucket, error) {
	resp, err := b.makeAndDoBucketRequest("/b2api/v1/b2_list_buckets", "", "", "")
	if err != nil {
		return nil, err
	}
	return b.listBuckets(resp)
}

func (b *B2) listBuckets(resp *http.Response) ([]Bucket, error) {
	defer resp.Body.Close()
	respBody := &listBucketsResponse{}
	err := ParseResponse(resp, respBody)
	if err != nil {
		return nil, err
	}

	for i := range respBody.Buckets {
		respBody.Buckets[i].B2 = b
	}
	return respBody.Buckets, nil
}

func (b *B2) CreateBucket(name string, bType BucketType) (*Bucket, error) {
	resp, err := b.makeAndDoBucketRequest("/b2api/v1/b2_list_buckets", "", name, bType)
	if err != nil {
		return nil, err
	}
	return b.createBucket(resp)
}

func (b *B2) createBucket(resp *http.Response) (*Bucket, error) {
	defer resp.Body.Close()
	bucket := &Bucket{B2: b}
	err := ParseResponse(resp, bucket)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (b *Bucket) Update(newBucketType BucketType) error {
	resp, err := b.B2.makeAndDoBucketRequest("/b2api/v1/b2_update_bucket", b.BucketID, "", newBucketType)
	if err != nil {
		return err
	}
	return b.update(resp)
}

func (b *Bucket) update(resp *http.Response) error {
	defer resp.Body.Close()
	return ParseResponse(resp, b)
}

func (b *Bucket) Delete() error {
	resp, err := b.B2.makeAndDoBucketRequest("/b2api/v1/b2_delete_bucket", b.BucketID, "", "")
	if err != nil {
		return err
	}
	return b.bucketDelete(resp)
}

func (b *Bucket) bucketDelete(resp *http.Response) error {
	defer resp.Body.Close()
	return ParseResponse(resp, b)
}

func (b *B2) makeAndDoBucketRequest(path, bucketID, bucketName string, bucketType BucketType) (*http.Response, error) {
	req, err := b.makeBucketRequest(path, bucketID, bucketName, bucketType)
	if err != nil {
		return nil, err
	}
	return b.client.Do(req)
}

func (b *B2) makeBucketRequest(path, bucketID, bucketName string, bucketType BucketType) (*http.Request, error) {
	requestBody := bucketRequest{
		AccountID:  b.AccountID,
		BucketID:   bucketID,
		BucketName: bucketName,
		BucketType: bucketType,
	}
	req, err := b.CreateRequest("POST", b.ApiUrl+path, requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.AuthorizationToken)
	return req, nil
}
