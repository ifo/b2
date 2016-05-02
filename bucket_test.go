package b2

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestB2_ListBuckets(t *testing.T) {
	b2 := testB2()
	b2.ListBuckets()
	req := b2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != b2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", b2.AuthorizationToken, auth)
	}
}

func TestB2_parseListBuckets(t *testing.T) {
	resp := testResponse(200, `{"buckets":[{"bucketId":"id","accountId":"id","bucketName":"name","bucketType":"allPrivate"}]}`)
	b2 := &B2{}
	buckets, err := b2.parseListBuckets(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(buckets) != 1 {
		t.Errorf("Expected one bucket, instead got %d", len(buckets))
	}
	if buckets[0].ID != "id" {
		t.Errorf(`Expected "id", instead got %s`, buckets[0].ID)
	}
	if buckets[0].Name != "name" {
		t.Errorf(`Expected "name", instead got %s`, buckets[0].Name)
	}
	if buckets[0].Type != AllPrivate {
		t.Errorf("Expected AllPrivate, instead got %+v", buckets[0].Type)
	}
	if *buckets[0].B2 != *b2 {
		t.Errorf("Expected bucket B2 to be *b2, instead got %+v", *buckets[0].B2)
	}

	resps := testResponseErrors()
	for i, resp := range resps {
		buckets, err := b2.parseListBuckets(resp)
		checkResponseError(err, 400+i, t)
		if buckets != nil {
			t.Errorf("Expected b2 to be nil, instead got %+v", b2)
		}
	}
}

func TestB2_CreateBucket(t *testing.T) {
	b2 := testB2()
	b2.CreateBucket("name", AllPrivate)
	req := b2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != b2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", b2.AuthorizationToken, auth)
	}
}

func TestB2_parseCreateBucket(t *testing.T) {
	resp := testResponse(200, `{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPrivate"}`)
	b2 := &B2{}
	bucket, err := b2.parseCreateBucket(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if bucket.ID != "id" {
		t.Errorf(`Expected "id", instead got %s`, bucket.ID)
	}
	if bucket.Name != "bucket" {
		t.Errorf(`Expected "bucket", instead got %s`, bucket.Name)
	}
	if bucket.Type != AllPrivate {
		t.Errorf("Expected bucket type to be private, instead got %s", bucket.Type)
	}
	if bucket.B2 != b2 {
		t.Errorf("Expected bucket B2 to be test B2, instead got %+v", bucket.B2)
	}

	resps := testResponseErrors()
	for i, resp := range resps {
		buckets, err := b2.parseCreateBucket(resp)
		checkResponseError(err, 400+i, t)
		if buckets != nil {
			t.Errorf("Expected b2 to be nil, instead got %+v", b2)
		}
	}
}

func TestBucket_Update(t *testing.T) {
	bucket := testBucket()
	bucket.Update(AllPrivate)
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_parseUpdate(t *testing.T) {
	resp := testResponse(200, `{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)
	bucket := testBucket()
	err := bucket.parseUpdate(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	// bucket type should change
	if bucket.Type != AllPublic {
		t.Errorf("Expected bucket type to be private, instead got %s", bucket.Type)
	}

	// nothing else should have changed
	if bucket.ID != "id" {
		t.Errorf(`Expected "id", instead got %s`, bucket.ID)
	}
	if bucket.Name != "bucket" {
		t.Errorf(`Expected "bucket", instead got %s`, bucket.Name)
	}

	resps := testResponseErrors()
	for i, resp := range resps {
		bucket := testBucket()
		err := bucket.parseUpdate(resp)
		checkResponseError(err, 400+i, t)
		if bucket.Type != AllPrivate {
			t.Errorf("Expected bucket type to be private, instead got %+v", bucket.Type)
		}
	}
}

func TestBucket_Delete(t *testing.T) {
	bucket := testBucket()
	bucket.Delete()
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_parseDelete(t *testing.T) {
	resp := testResponse(200, `{"bucketId":"id","accountId":"id","bucketName":"bucket","bucketType":"allPublic"}`)
	bucket := testBucket()
	err := bucket.parseDelete(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	resps := testResponseErrors()
	for i, resp := range resps {
		bucket := testBucket()
		err := bucket.parseDelete(resp)
		checkResponseError(err, 400+i, t)
	}
}

func TestB2_createBucketRequest(t *testing.T) {
	b2 := &B2{APIURL: "http://example.com"}
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
	req1, err := b2.createBucketRequest("/", brs[0])
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body1, err := ioutil.ReadAll(req1.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body1)

	req2, err := b2.createBucketRequest("/", brs[1])
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	body2, err := ioutil.ReadAll(req2.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body2)

	req3, err := b2.createBucketRequest("/", brs[2])
	if err != nil {
		t.Errorf("Expected no error, instead got %s", err)
	}
	body3, err := ioutil.ReadAll(req3.Body)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	reqs = append(reqs, body3)

	req4, err := b2.createBucketRequest("/", brs[3])
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

func testBucket() *Bucket {
	return &Bucket{
		ID:   "id",
		Name: "bucket",
		Type: AllPrivate,
		B2:   testB2(),
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
