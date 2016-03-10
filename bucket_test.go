package b2

import (
	"testing"
)

func Test_B2_ListBuckets_Success(t *testing.T) {
	b := makeTestB2()
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

func Test_B2_ListBuckets_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		buckets, err := b.ListBuckets()
		testErrorResponse(err, codes[i], t)
		if buckets != nil {
			t.Errorf("Expected buckets to be empty, instead got %+v", buckets)
		}

		s.Close()
	}
}
