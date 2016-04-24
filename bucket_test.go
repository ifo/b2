package b2

import (
	"bytes"
	"io/ioutil"
	"testing"

	"net/http" // remove soon with refactor
)

func Test_listBuckets(t *testing.T) {
	resp := createTestResponse(200, `{"buckets":
[{"bucketId":"id","accountId":"id","bucketName":"name","bucketType":"allPrivate"}]}`)

	b := &B2{}
	buckets, err := b.listBuckets(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(buckets) != 1 {
		t.Errorf("Expected one bucket, instead got %d", len(buckets))
	}
	if buckets[0].BucketID != "id" {
		t.Errorf(`Expected "id", instead got %s`, buckets[0].BucketID)
	}
	if buckets[0].BucketName != "name" {
		t.Errorf(`Expected "name", instead got %s`, buckets[0].BucketName)
	}
	if buckets[0].BucketType != AllPrivate {
		t.Errorf("Expected AllPrivate, instead got %+v", buckets[0].BucketType)
	}
	if *buckets[0].B2 != *b {
		t.Errorf("Expected bucket B2 to be *b, instead got %+v", *buckets[0].B2)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		buckets, err := b.listBuckets(resp)
		testErrorResponse(err, 400+i, t)
		if buckets != nil {
			t.Errorf("Expected b to be nil, instead got %+v", b)
		}
	}
}

func Test_B2_createBucket(t *testing.T) {
	resp := createTestResponse(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPrivate"}`)

	b := &B2{}
	bucket, err := b.createBucket(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	if bucket.BucketID != "id" {
		t.Errorf(`Expected "id", instead got %s`, bucket.BucketID)
	}
	if bucket.BucketName != "bucket" {
		t.Errorf(`Expected "bucket", instead got %s`, bucket.BucketName)
	}
	if bucket.BucketType != AllPrivate {
		t.Errorf("Expected bucket type to be private, instead got %s", bucket.BucketType)
	}
	if bucket.B2 != b {
		t.Errorf("Expected bucket B2 to be test B2, instead got %+v", bucket.B2)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		buckets, err := b.listBuckets(resp)
		testErrorResponse(err, 400+i, t)
		if buckets != nil {
			t.Errorf("Expected b to be nil, instead got %+v", b)
		}
	}
}

func Test_Bucket_update(t *testing.T) {
	resp := createTestResponse(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)

	b := &B2{}
	bucket := makeTestBucket(b)
	err := bucket.update(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	// bucket type should change
	if bucket.BucketType != AllPublic {
		t.Errorf("Expected bucket type to be private, instead got %s", bucket.BucketType)
	}

	// nothing else should have changed
	if bucket.BucketID != "id" {
		t.Errorf(`Expected "id", instead got %s`, bucket.BucketID)
	}
	if bucket.BucketName != "bucket" {
		t.Errorf(`Expected "bucket", instead got %s`, bucket.BucketName)
	}
	if bucket.B2 != b {
		t.Errorf("Expected bucket B2 to be test B2, instead got %+v", bucket.B2)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		bucket := makeTestBucket(b)
		err := bucket.update(resp)
		testErrorResponse(err, 400+i, t)
		if bucket.BucketType != AllPrivate {
			t.Errorf("Expected bucket type to be private, instead got %+v", bucket.BucketType)
		}
	}
}

func Test_Bucket_bucketDelete(t *testing.T) {
	resp := createTestResponse(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)

	bucket := makeTestBucket(&B2{})
	err := bucket.bucketDelete(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		bucket := makeTestBucket(&B2{})
		err := bucket.bucketDelete(resp)
		testErrorResponse(err, 400+i, t)
	}
}

func Test_B2_createBucketRequest(t *testing.T) {
	b := makeTestB2(http.Client{})

	reqs := [][]byte{}
	fields := [][]byte{
		[]byte("accountId"), []byte("bucketId"), []byte("bucketName"), []byte("bucketType")}
	brs := []bucketRequest{
		{}, // List Buckets
		{BucketName: "bucket", BucketType: AllPublic}, // Create Bucket
		{BucketID: "id", BucketType: AllPrivate},      // Bucket Update
		{BucketID: "id"},                              // Bucket Delete
	}
	finds := [][][]byte{
		{fields[0]},                       // List Buckets
		{fields[0], fields[2], fields[3]}, // Create Bucket
		{fields[0], fields[1], fields[3]}, // Bucket Update
		{fields[0], fields[1]},            // Bucket Delete
	}

	// Setup all request bodies
	req1, err := b.createBucketRequest("/", brs[0])
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body1, err := ioutil.ReadAll(req1.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body1)

	req2, err := b.createBucketRequest("/", brs[1])
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body2, err := ioutil.ReadAll(req2.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body2)

	req3, err := b.createBucketRequest("/", brs[2])
	if err != nil {
		t.Errorf("Expected no error, instead got %s", err)
	}
	body3, err := ioutil.ReadAll(req3.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body3)

	req4, err := b.createBucketRequest("/", brs[3])
	if err != nil {
		t.Errorf("Expected no error, instead got %s", err)
	}
	body4, err := ioutil.ReadAll(req4.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body4)

	for i, reqBody := range reqs {
		for _, elem := range fields {
			// should have but doesn't
			if sliceContains(finds[i], elem) && !bytes.Contains(reqBody, elem) {
				t.Errorf("reqBody should have %q, but only had %q", elem, reqBody)
			}
			// shouldn't have but does
			if !sliceContains(finds[i], elem) && bytes.Contains(reqBody, elem) {
				t.Errorf("reqBody should not have %q, but does, and has %q", elem, reqBody)
			}
		}
	}
}

func makeTestBucket(b *B2) *Bucket {
	return &Bucket{
		BucketID:   "id",
		BucketName: "bucket",
		BucketType: AllPrivate,
		B2:         b,
	}
}

func sliceContains(haystack [][]byte, needle []byte) bool {
	for _, straw := range haystack {
		if bytes.Equal(straw, needle) {
			return true
		}
	}
	return false
}
