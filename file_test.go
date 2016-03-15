package b2

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_Bucket_ListFileNames_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)

	fileAction := []Action{ActionUpload, ActionUpload, ActionUpload}
	setupFiles := ""
	for i := range fileAction {
		setupFiles += makeTestFileJson(i, fileAction[i])
		if i != len(fileAction)-1 {
			setupFiles += ","
		}
	}
	s := setupRequest(200, fmt.Sprintf(`{"files":[%s],"nextFileName":"name%d"}`, setupFiles, len(fileAction)))
	defer s.Close()

	response, err := bucket.ListFileNames("", 3)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(response.Files) != 3 {
		t.Fatalf("Expected two files, instead got %d", len(response.Files))
	}
	if response.NextFileName != fmt.Sprintf("name%d", len(fileAction)) {
		t.Errorf("Expected next file name to be name%d, instead got %s", len(fileAction), response.NextFileName)
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
		if file.UploadTimestamp != int64(100+i) {
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

func Test_Bucket_ListFileVersions_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)

	fileAction := []Action{ActionUpload, ActionHide, ActionStart}
	setupFiles := ""
	for i := range fileAction {
		setupFiles += makeTestFileJson(i, fileAction[i])
		if i != len(fileAction)-1 {
			setupFiles += ","
		}
	}
	s := setupRequest(200, fmt.Sprintf(`{"files":[%s],"nextFileId":"id%d","nextFileName":"name%d"}`,
		setupFiles, len(fileAction), len(fileAction)))
	defer s.Close()

	response, err := bucket.ListFileVersions("", "", 3)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(response.Files) != 3 {
		t.Fatalf("Expected three files, instead got %d", len(response.Files))
	}
	if response.NextFileName != "name3" {
		t.Errorf("Expected next file name to be name3, instead got %s", response.NextFileName)
	}
	if response.NextFileID != "id3" {
		t.Errorf("Expected next file id to be id3, instead got %s", response.NextFileID)
	}
	for i, file := range response.Files {
		if file.Action != fileAction[i] {
			t.Errorf("Expected action to be %v, instead got %v", fileAction[i], file.Action)
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
		if file.UploadTimestamp != int64(100+i) {
			t.Errorf("Expected upload timestamp to be %d, instead got %d", 10+i, file.UploadTimestamp)
		}
		if file.Bucket != bucket {
			t.Errorf("Expected file bucket to be bucket, instead got %+v", file.Bucket)
		}
	}
}

func Test_Bucket_ListFileVersions_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		response, err := bucket.ListFileVersions("", "", 0)
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}
}

func Test_Bucket_GetFileInfo_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)

	fileAction := []Action{ActionUpload, ActionHide, ActionStart}

	for i := range fileAction {
		s := setupRequest(200, makeTestFileJson(i, fileAction[i]))

		fileID := fmt.Sprintf("id%d", i)
		fileMeta, err := bucket.GetFileInfo(fileID)
		if err != nil {
			t.Fatalf("Expected no error, instead got %s", err)
		}

		if fileMeta.Action != fileAction[i] {
			t.Errorf("Expected action to be %v, instead got %v", fileAction[i], fileMeta.Action)
		}
		if fileMeta.ID != fmt.Sprintf("id%d", i) {
			t.Errorf("Expected file ID to be id%d, instead got %s", i, fileMeta.ID)
		}
		if fileMeta.Name != fmt.Sprintf("name%d", i) {
			t.Errorf("Expected file name to be name%d, instead got %s", i, fileMeta.Name)
		}
		if fileMeta.ContentLength != int64(10+i) {
			t.Errorf("Expected content length to be %d, instead got %d", 10+i, fileMeta.ContentLength)
		}
		if fileMeta.ContentSha1 != "sha1" {
			t.Errorf(`Expected content sha1 to be "sha1", instead got %s`, fileMeta.ContentSha1)
		}
		if fileMeta.ContentType != "text" {
			t.Errorf("Expected content type to be text, instead got %s", fileMeta.ContentType)
		}
		if fileMeta.Bucket != bucket {
			t.Errorf("Expected file bucket to be bucket, instead got %+v", fileMeta.Bucket)
		}
		for k, v := range fileMeta.FileInfo {
			t.Errorf("Expected fileInfo to be blank, instead got %s, %s", k, v)
		}

		s.Close()
	}
}

func Test_Bucket_GetFileInfo_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		response, err := bucket.GetFileInfo(fmt.Sprintf("id%d", i))
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}

	// test no provided ID error
	response, err := bucket.GetFileInfo("")
	if err.Error() != "No fileID provided" {
		t.Errorf(`Expected "No fileID provided", instead got %s`, err)
	}
	if response != nil {
		t.Errorf("Expected response to be empty, instead got %+v", response)
	}
}

func Test_Bucket_UploadFile_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)
	bucket.UploadUrls = append(bucket.UploadUrls, makeTestUploadUrl())

	s := setupRequest(200, makeTestFileJson(0, ActionUpload))
	defer s.Close()

	// TODO test fileInfo
	fileMeta, err := bucket.UploadFile("name0", strings.NewReader("length ten"), nil)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if fileMeta.Action != ActionUpload {
		t.Errorf("Expected action to be upload, instead got %v", fileMeta.Action)
	}
	if fileMeta.ID != "id0" {
		t.Errorf("Expected file ID to be id0, instead got %s", fileMeta.ID)
	}
	if fileMeta.Name != "name0" {
		t.Errorf("Expected file name to be name0, instead got %s", fileMeta.Name)
	}
	if fileMeta.ContentLength != 10 {
		t.Errorf("Expected content length to be 10, instead got %d", fileMeta.ContentLength)
	}
	if fileMeta.ContentSha1 != "sha1" {
		t.Errorf(`Expected content sha1 to be "sha1", instead got %s`, fileMeta.ContentSha1)
	}
	if fileMeta.ContentType != "text" {
		t.Errorf("Expected content type to be text/plain, instead got %s", fileMeta.ContentType)
	}
	if fileMeta.Bucket != bucket {
		t.Errorf("Expected file bucket to be bucket, instead got %+v", fileMeta.Bucket)
	}
	for k, v := range fileMeta.FileInfo {
		t.Errorf("Expected fileInfo to be blank, instead got %s, %s", k, v)
	}
}

func Test_Bucket_UploadFile_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		response, err := bucket.UploadFile("", strings.NewReader(""), nil)
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}
}

func Test_Bucket_GetUploadUrl_Success(t *testing.T) {
	b := makeTestB2()
	bucket := makeTestBucket(b)

	uploadUrl := "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id"

	s := setupRequest(200, fmt.Sprintf(`{"bucketId":"id","uploadUrl":"%s","authorizationToken":"token"}`, uploadUrl))
	defer s.Close()

	response, err := bucket.GetUploadUrl()
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if response.Expiration.IsZero() {
		t.Error("Expected time to be now + 24h, instead got zero time")
	}
	if response.AuthorizationToken != "token" {
		t.Errorf(`Expected response token to be "token", instead got %s`, response.AuthorizationToken)
	}
	if response.Url != uploadUrl {
		t.Errorf("Expected response url to be uploadUrl, instead got %s", response.Url)
	}

	if len(bucket.UploadUrls) != 1 {
		t.Fatalf("Expected length of bucket upload urls to be 1, insetad was %d", len(bucket.UploadUrls))
	}
	if bucket.UploadUrls[0] != response {
		t.Error("Expected bucket's uploadUrls to be response, instead was", bucket.UploadUrls[0])
	}
}

func Test_Bucket_GetUploadUrl_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	b := makeTestB2()
	bucket := makeTestBucket(b)

	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		response, err := bucket.GetUploadUrl()
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}
		if len(bucket.UploadUrls) != 0 {
			t.Errorf("Expected no upload urls, instead got %d", bucket.UploadUrls)
		}

		s.Close()
	}
}

func Test_Bucket_cleanUploadUrls(t *testing.T) {
	t.Skip()
}

func makeTestFileJson(num int, action Action) string {
	file := FileMeta{
		ID:              fmt.Sprintf("id%d", num),
		Name:            fmt.Sprintf("name%d", num),
		Size:            int64(10 + num),
		ContentLength:   int64(10 + num),
		ContentSha1:     "sha1", // TODO make valid SHA1
		ContentType:     "text",
		Action:          action,
		FileInfo:        map[string]string{},
		UploadTimestamp: int64(100 + num),
	}
	fileJson, _ := json.Marshal(file)
	return string(fileJson)
}

func makeTestUploadUrl() *UploadUrl {
	return &UploadUrl{
		Url:                "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id",
		AuthorizationToken: "token",
		Expiration:         time.Now().UTC().Add(24 * time.Hour),
	}
}
