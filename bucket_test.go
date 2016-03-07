package b2

import (
	"testing"
)

func Test_ListBuckets_200(t *testing.T) {
	b := &B2{
		AccountID:          "1",
		AuthorizationToken: "1",
		ApiUrl:             "https://api001.backblaze.com",
	}

	s := setupRequest(200, `{"buckets": [{
    "bucketId": "id",
    "accountId": "id",
    "bucketName" : "name",
    "bucketType": "allPrivate"}]}`)
	defer s.Close()

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

func Test_ListBuckets_400(t *testing.T) {
	b := &B2{
		AccountID:          "1",
		AuthorizationToken: "1",
		ApiUrl:             "https://api001.backblaze.com",
	}

	s := setupRequest(400,
		`{"status":400,"code":"nope","message":"nope nope"}`)
	defer s.Close()

	buckets, err := b.ListBuckets()
	if err == nil {
		t.Fatal("Expected error, no error received")
	}
	if err.Error() != "Status: 400, Code: nope, Message: nope nope" {
		t.Errorf(`Expected "Status: 400, Code: nope, Message: nope nope", instead got %s`, err)
	}
	if buckets != nil {
		t.Errorf("Expected b to be empty, instead got %+v", buckets)
	}
}

func Test_ListBuckets_401(t *testing.T) {
	b := &B2{
		AccountID:          "1",
		AuthorizationToken: "1",
		ApiUrl:             "https://api001.backblaze.com",
	}

	s := setupRequest(401,
		`{"status":401,"code":"nope","message":"nope nope"}`)
	defer s.Close()

	buckets, err := b.ListBuckets()
	if err == nil {
		t.Fatal("Expected error, no error received")
	}
	if err.Error() != "Status: 401, Code: nope, Message: nope nope" {
		t.Errorf(`Expected "Status: 401, Code: nope, Message: nope nope", instead got %s`, err)
	}
	if buckets != nil {
		t.Errorf("Expected buckets to be empty, instead got %+v", buckets)
	}
}
