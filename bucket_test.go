package b2

import (
	"testing"
)

func Test_B2_ListBuckets_Success(t *testing.T) {
	b := makeTestB2()
	s := setupRequest(200, `{"buckets":
[{"bucketId": "id","accountId": "id","bucketName" : "name","bucketType": "allPrivate"}]}`)
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

func Test_B2_CreateBucket_Success(t *testing.T) {
	b := makeTestB2()
	s := setupRequest(200,
		`{"bucketId": "id","accountId": "id","bucketName" : "bucket","bucketType": "allPrivate"}`)
	defer s.Close()

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
	b := makeTestB2()

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		bucket, err := b.CreateBucket("bucket", AllPrivate)
		testErrorResponse(err, codes[i], t)
		if bucket != nil {
			t.Errorf("Expected bucket to be empty, instead got %+v", bucket)
		}

		s.Close()
	}
}

func Test_Bucket_Update_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)
	s := setupRequest(200,
		`{"bucketId": "id","accountId": "id","bucketName" : "bucket","bucketType": "allPublic"}`)
	defer s.Close()

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
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		err := bucket.Update(AllPublic)
		testErrorResponse(err, codes[i], t)
		if bucket.BucketType != AllPrivate {
			t.Errorf("Expected bucket type to be private, instead got %+v", bucket.BucketType)
		}

		s.Close()
	}
}

func Test_Bucket_Delete_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)
	s := setupRequest(200,
		`{"bucketId": "id","accountId": "id","bucketName" : "bucket","bucketType": "allPublic"}`)
	defer s.Close()

	err := bucket.Delete()
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
}

func Test_Bucket_Delete_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		err := bucket.Update(AllPublic)
		testErrorResponse(err, codes[i], t)

		s.Close()
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
