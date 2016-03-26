package b2

import (
	"bytes"
	"io/ioutil"
	"testing"

	"net/http" // remove soon with refactor
)

func Test_B2_ListBuckets_Success(t *testing.T) {
	s, c := setupRequest(200, `{"buckets":
[{"bucketId":"id","accountId":"id","bucketName":"name","bucketType":"allPrivate"}]}`)
	defer s.Close()

	b := makeTestB2(c)
	buckets, err := b.ListBuckets()
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
}

func Test_B2_ListBuckets_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		buckets, err := b.ListBuckets()
		testErrorResponse(err, codes[i], t)
		if buckets != nil {
			t.Errorf("Expected buckets to be empty, instead got %+v", buckets)
		}

		s.Close()
	}
}

func Test_B2_CreateBucket_Success(t *testing.T) {
	s, c := setupRequest(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPrivate"}`)
	defer s.Close()

	b := makeTestB2(c)
	bucket, err := b.CreateBucket("bucket", AllPrivate)
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
		t.Errorf("Expected bucket B2 to be test B2, instead got %s", bucket.B2)
	}
}

func Test_B2_CreateBucket_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket, err := b.CreateBucket("bucket", AllPrivate)
		testErrorResponse(err, codes[i], t)
		if bucket != nil {
			t.Errorf("Expected bucket to be empty, instead got %+v", bucket)
		}

		s.Close()
	}
}

func Test_Bucket_Update_Success(t *testing.T) {
	s, c := setupRequest(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	err := bucket.Update(AllPublic)
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
		t.Errorf("Expected bucket B2 to be test B2, instead got %s", bucket.B2)
	}
}

func Test_Bucket_Update_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		err := bucket.Update(AllPublic)
		testErrorResponse(err, codes[i], t)
		if bucket.BucketType != AllPrivate {
			t.Errorf("Expected bucket type to be private, instead got %+v", bucket.BucketType)
		}

		s.Close()
	}
}

func Test_Bucket_Delete_Success(t *testing.T) {
	s, c := setupRequest(200,
		`{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	err := bucket.Delete()
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
}

func Test_Bucket_Delete_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		err := bucket.Update(AllPublic)
		testErrorResponse(err, codes[i], t)

		s.Close()
	}
}

func Test_B2_makeBucketRequest(t *testing.T) {
	b := makeTestB2(http.Client{})

	reqs := [][]byte{}
	fields := [][]byte{
		[]byte("accountId"), []byte("bucketId"), []byte("bucketName"), []byte("bucketType")}
	finds := [][][]byte{
		[][]byte{fields[0]},                       // List Buckets
		[][]byte{fields[0], fields[2], fields[3]}, // Create Bucket
		[][]byte{fields[0], fields[1], fields[3]}, // Bucket Update
		[][]byte{fields[0], fields[1]},            // Bucket Delete
	}

	// Setup all request bodies
	req1, err := b.makeBucketRequest("/", "", "", "")
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body1, err := ioutil.ReadAll(req1.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body1)

	req2, err := b.makeBucketRequest("/", "", "name", AllPublic)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body2, err := ioutil.ReadAll(req2.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body2)

	req3, err := b.makeBucketRequest("/", "id", "", AllPrivate)
	if err != nil {
		t.Errorf("Expected no error, instead got %s", err)
	}
	body3, err := ioutil.ReadAll(req3.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body3)

	req4, err := b.makeBucketRequest("/", "id", "", "")
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
