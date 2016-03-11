package b2

import (
	"fmt"
	"testing"
)

func Test_Bucket_ListFileNames_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)
	s := setupRequest(200, `{"files":[
{"action":"upload","fileId":"id0","fileName":"name0","size":10,"uploadTimestamp":10},
{"action":"upload","fileId":"id1","fileName":"name1","size":11,"uploadTimestamp":11}],
"nextFileName":null}`)
	defer s.Close()

	response, err := bucket.ListFileNames("", 0)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(response.Files) != 2 {
		t.Fatalf("Expected two files, instead got %d", len(response.Files))
	}
	if response.NextFileName != "" {
		t.Errorf("Expected no next file name, instead got %s", response.NextFileName)
	}
	if response.NextFileID != "" {
		t.Errorf("Expected no next file id, instead got %s", response.NextFileID)
	}
	for i, file := range response.Files {
		if file.Action != ActionUpload {
			t.Errorf("Expected action to be upload, instead got %v", file.Action)
		}
		if file.ID != fmt.Sprintf("id%d", i) {
			t.Errorf("Expected file ID to be id%d, instead got %s", i, fmt.Sprintf("id%d", i))
		}
		if file.Name != fmt.Sprintf("name%d", i) {
			t.Errorf("Expected file name to be name%d, instead got %s", i, fmt.Sprintf("name%d", i))
		}
		if file.Size != int64(10+i) {
			t.Errorf("Expected size to be %d, instead got %d", 10+i, file.Size)
		}
		if file.UploadTimestamp != int64(10+i) {
			t.Errorf("Expected upload timestamp to be %d, instead got %d", 10+i, file.UploadTimestamp)
		}
		if file.Bucket != bucket {
			t.Errorf("Expected file bucket to be bucket, instead got %+v", file.Bucket)
		}
	}
}

func Test_Bucket_ListFileNames_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		response, err := bucket.ListFileNames("", 0)
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}
}
