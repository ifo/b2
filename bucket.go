package b2

import ()

type Bucket struct {
	BucketID   string
	BucketName string
	BucketType BucketType
	B2         *B2
}

type BucketType string

const (
	AllPrivate BucketType = "allPrivate"
	AllPublic  BucketType = "allPublic"
)

func (b *B2) ListBuckets() ([]Bucket, error) {
	return []Bucket{}, nil
}

func (b *B2) CreateBucket(name string, bType BucketType) (*Bucket, error) {
	return &Bucket{}, nil
}

func (b *Bucket) Update(bType BucketType) (*Bucket, error) {
	return &Bucket{}, nil
}

func (b *Bucket) Delete() error {
	return nil
}
