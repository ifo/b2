package b2

import (
	"net/http"
	"time"
)

// Bucket contains all data about a B2 Bucket. It also has a reference to
// the B2 account which it is under.
type Bucket struct {
	ID         string       `json:"bucketId"`
	Name       string       `json:"bucketName"`
	Type       BucketType   `json:"bucketType"`
	UploadURLs []*UploadURL `json:"-"`
	B2         *B2          `json:"-"`
}

// BucketType is the visibility of a bucket.
type BucketType string

// BucketTypes can be private or public.
const (
	AllPrivate BucketType = "allPrivate"
	AllPublic  BucketType = "allPublic"
)

// UploadURL is a special URL used for upolading files to a bucket. It has
// its own separate Authorization Token, and expires 24 hours after creation.
type UploadURL struct {
	URL                string    `json:"uploadUrl"`
	AuthorizationToken string    `json:"authorizationToken"`
	Expiration         time.Time `json:"-"`
	// TODO? include some marker of being used (maybe mutex)
}

type listBucketsResponse struct {
	Buckets []Bucket `json:"buckets"`
}

// bucketRequest is used for making any bucket related request.
// The accountID is always required, though the other parameters vary.
type bucketRequest struct {
	AccountID  string     `json:"accountId"`
	BucketID   string     `json:"bucketId,omitempty"`
	BucketName string     `json:"bucketName,omitempty"`
	BucketType BucketType `json:"bucketType,omitempty"`
}

// ListBuckets gets a list of all buckets in an account.
//
// It also sets up the necessary reference to the B2 API client.
func (b2 *B2) ListBuckets() ([]Bucket, error) {
	req, err := b2.createBucketRequest("/b2api/v1/b2_list_buckets", bucketRequest{})
	if err != nil {
		return nil, err
	}

	resp, err := b2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b2.parseListBuckets(resp)
}

func (b2 *B2) parseListBuckets(resp *http.Response) ([]Bucket, error) {
	lbr := &listBucketsResponse{}
	err := parseResponse(resp, lbr)
	if err != nil {
		return nil, err
	}

	for i := range lbr.Buckets {
		lbr.Buckets[i].B2 = b2
	}
	return lbr.Buckets, nil
}

// CreateBucket creates a new bucket with the given name and type.
func (b2 *B2) CreateBucket(name string, t BucketType) (*Bucket, error) {
	br := bucketRequest{BucketName: name, BucketType: t}
	req, err := b2.createBucketRequest("/b2api/v1/b2_list_buckets", br)
	if err != nil {
		return nil, err
	}

	resp, err := b2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b2.parseCreateBucket(resp)
}

func (b2 *B2) parseCreateBucket(resp *http.Response) (*Bucket, error) {
	b := &Bucket{B2: b2}
	err := parseResponse(resp, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Update sets the bucket type on a bucket.
//
// The type is modified in place if successful, and unchanged otherwise.
func (b *Bucket) Update(newBucketType BucketType) error {
	br := bucketRequest{BucketID: b.ID, BucketType: newBucketType}
	req, err := b.B2.createBucketRequest("/b2api/v1/b2_update_bucket", br)
	if err != nil {
		return err
	}

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return err
	}
	return b.parseUpdate(resp)
}

func (b *Bucket) parseUpdate(resp *http.Response) error {
	return parseResponse(resp, b)
}

// Delete removes a bucket. The bucket reference itself is unchanged.
func (b *Bucket) Delete() error {
	br := bucketRequest{BucketID: b.ID}
	req, err := b.B2.createBucketRequest("/b2api/v1/b2_delete_bucket", br)
	if err != nil {
		return err
	}

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return err
	}
	return b.parseDelete(resp)
}

func (b *Bucket) parseDelete(resp *http.Response) error {
	return parseResponse(resp, b)
}

// createBucketRequest makes a http.Request from a given bucketRequest.
// It ensures that the bucketRequest defines the required AccountID.
func (b2 *B2) createBucketRequest(path string, br bucketRequest) (*http.Request, error) {
	br.AccountID = b2.AccountID
	req, err := CreateRequest("POST", b2.APIURL+path, br)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b2.AuthorizationToken)
	return req, nil
}
