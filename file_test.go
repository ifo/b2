package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func Test_Bucket_ListfileNames(t *testing.T) {
	bucket := createTestBucket()
	bucket.ListFileNames("name", 1)
	req := bucket.B2.client.(*dummyClient).Req
	auth := req.Header["Authorization"][0]
	if auth != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func Test_Bucket_parseListFile(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}
	setupFiles := ""
	for i := range fileAction {
		setupFiles += createTestFileJson(i, fileAction[i], nil)
		if i != len(fileAction)-1 {
			setupFiles += ","
		}
	}
	resp := createTestResponse(200, fmt.Sprintf(`{"files":[%s],"nextFileId":"id%d","nextFileName":"name%d"}`,
		setupFiles, len(fileAction), len(fileAction)))

	bucket := createTestBucket()
	fileList, err := bucket.parseListFile(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if len(fileList.Files) != 3 {
		t.Fatalf("Expected three files, instead got %d", len(fileList.Files))
	}
	if fileList.NextFileName != "name3" {
		t.Errorf("Expected next file name to be name3, instead got %s", fileList.NextFileName)
	}
	if fileList.NextFileID != "id3" {
		t.Errorf("Expected next file id to be id3, instead got %s", fileList.NextFileID)
	}
	for i, file := range fileList.Files {
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

	resps := createTestErrorResponses()
	for i, resp := range resps {
		fileList, err := bucket.parseListFile(resp)
		testErrorResponse(err, 400+i, t)
		if fileList != nil {
			t.Errorf("Expected fileList to be empty, instead got %+v", fileList)
		}
	}
}

func Test_Bucket_parseFileMetaResponse(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}

	for i := range fileAction {
		resp := createTestResponse(200, createTestFileJson(i, fileAction[i], nil))

		bucket := createTestBucket()

		fileMeta, err := bucket.parseFileMetaResponse(resp)
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
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		bucket := createTestBucket()
		fileMeta, err := bucket.parseFileMetaResponse(resp)
		testErrorResponse(err, 400+i, t)
		if fileMeta != nil {
			t.Errorf("Expected response to be empty, instead got %+v", fileMeta)
		}
	}
}

func Test_Bucket_GetFileInfo_NoFileID(t *testing.T) {
	bucket := createTestBucket()
	fileInfo, err := bucket.GetFileInfo("")
	if err.Error() != "No fileID provided" {
		t.Errorf(`Expected "No fileID provided", instead got %s`, err)
	}
	if fileInfo != nil {
		t.Errorf("Expected fileInfo to be empty, instead got %+v", fileInfo)
	}
}

func Test_Bucket_setupUploadFile(t *testing.T) {
	fileName := "cats.txt"
	fileData := bytes.NewReader([]byte("cats cats cats cats"))
	fileInfo := map[string]string{
		"file-cats": "yes",
		"file-dogs": "no",
	}

	fileInfoCheck := map[string]string{
		"Authorization":       "token2",
		"X-Bz-File-Name":      "cats.txt",
		"Content-Type":        "b2/x-auto",
		"Content-Length":      "19",
		"X-Bz-Content-Sha1":   "78498e5096b20e3f1c063e8740ff83d595ededb3",
		"X-Bz-Info-file-cats": fileInfo["file-cats"],
		"X-Bz-Info-file-dogs": fileInfo["file-dogs"],
	}

	uploadUrls := []*UploadUrl{
		{Url: "https://example.com/1", AuthorizationToken: "token1", Expiration: time.Now().UTC()}, // expired
		{Url: "https://example.com/2", AuthorizationToken: "token2", Expiration: time.Now().UTC().Add(1 * time.Hour)},
	}
	bucket := createTestBucket()
	bucket.UploadUrls = uploadUrls
	req, err := bucket.setupUploadFile(fileName, fileData, fileInfo)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	for k, v := range fileInfoCheck {
		if req.Header.Get(k) != v {
			t.Errorf("Expected req header %s to be %s, instead got %s", k, v, req.Header.Get(k))
		}
	}
}

func Test_Bucket_parseGetUploadUrlResponse(t *testing.T) {
	uploadUrlStr := "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id"
	resp := createTestResponse(200, fmt.Sprintf(`{"bucketId":"id","uploadUrl":"%s","authorizationToken":"token"}`, uploadUrlStr))

	bucket := createTestBucket()
	uploadUrl, err := bucket.parseGetUploadUrlResponse(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if uploadUrl.Expiration.IsZero() {
		t.Error("Expected time to be now + 24h, instead got zero time")
	}
	if uploadUrl.AuthorizationToken != "token" {
		t.Errorf(`Expected uploadUrl token to be "token", instead got %s`, uploadUrl.AuthorizationToken)
	}
	if uploadUrl.Url != uploadUrlStr {
		t.Errorf("Expected uploadUrl's url to be uploadUrlStr, instead got %s", uploadUrl.Url)
	}

	if len(bucket.UploadUrls) != 1 {
		t.Fatalf("Expected length of bucket upload urls to be 1, insetad was %d", len(bucket.UploadUrls))
	}
	if bucket.UploadUrls[0] != uploadUrl {
		t.Error("Expected bucket's first uploadUrl to be uploadUrl, instead was", bucket.UploadUrls[0])
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		bucket := createTestBucket()
		uploadUrl, err := bucket.parseGetUploadUrlResponse(resp)
		testErrorResponse(err, 400+i, t)
		if uploadUrl != nil {
			t.Errorf("Expected response to be empty, instead got %+v", uploadUrl)
		}
	}
}

func Test_Bucket_parseFileResponse(t *testing.T) {
	headers := map[string][]string{
		"X-Bz-File-Id":      {"1"},
		"X-Bz-File-Name":    {"cats.txt"},
		"Content-Length":    {"19"},
		"X-Bz-Content-Sha1": {"78498e5096b20e3f1c063e8740ff83d595ededb3"},
		"Content-Type":      {"text/plain"},
	}
	fileData := "cats cats cats cats"
	resp := createTestResponse(200, fileData)
	resp.Header = headers

	bucket := createTestBucket()
	file, err := bucket.parseFileResponse(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if file.Meta.ID != "1" {
		t.Errorf(`Expected file.Meta.ID to be "1", instead got %s`, file.Meta.ID)
	}
	if file.Meta.Name != "cats.txt" {
		t.Errorf(`Expected file.Meta.Name to be "cats.txt", instead got %s`, file.Meta.Name)
	}
	if file.Meta.Size != int64(len(file.Data)) {
		t.Errorf("Expected file.Meta.Size to be 19, instead got %d", file.Meta.Size)
	}
	if file.Meta.ContentLength != 19 {
		t.Errorf("Expected file.Meta.ContentLength to be 19, instead got %d", file.Meta.ContentLength)
	}
	if file.Meta.ContentSha1 != headers["X-Bz-Content-Sha1"][0] {
		t.Errorf(`Expected file.Meta.Sha1 to be "%s", instead got %s`, headers["X-Bz-Content-Sha1"], file.Meta.ContentSha1)
	}
	if file.Meta.ContentType != "text/plain" {
		t.Errorf(`Expected file.Meta.ContentType to be "text/plain", instead got %s`, file.Meta.ContentType)
	}
	// TODO include and test fileinfo
	for k, v := range file.Meta.FileInfo {
		t.Errorf("Expected fileInfo to be blank, instead got %s, %s", k, v)
	}
	if !bytes.Equal(file.Data, []byte(fileData)) {
		t.Errorf(`Expected file.Data to be "%v", instead got %v`, []byte(fileData), file.Data)
	}
	if file.Meta.Bucket != bucket {
		t.Errorf("Expected file.Meta.bucket to be bucket, instead got %+v", file.Meta.Bucket)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		bucket := createTestBucket()
		uploadUrl, err := bucket.parseFileResponse(resp)
		testErrorResponse(err, 400+i, t)
		if uploadUrl != nil {
			t.Errorf("Expected response to be empty, instead got %+v", uploadUrl)
		}
	}
}

func Test_Bucket_cleanUploadUrls(t *testing.T) {
	bucket := createTestBucket()

	times := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(1 * time.Hour),
		time.Now().UTC().Add(-1 * time.Hour),
		time.Now().UTC().Add(2 * time.Hour),
	}
	// two UploadUrls should be cleaned
	bucket.UploadUrls = append(bucket.UploadUrls, &UploadUrl{Expiration: times[0]})
	bucket.UploadUrls = append(bucket.UploadUrls, &UploadUrl{Expiration: times[1]})
	bucket.UploadUrls = append(bucket.UploadUrls, &UploadUrl{Expiration: times[2]})
	bucket.UploadUrls = append(bucket.UploadUrls, &UploadUrl{Expiration: times[3]})

	bucket.cleanUploadUrls()

	if len(bucket.UploadUrls) != 2 {
		t.Fatalf("Expected UploadUrls length to be 2, instead got %d", len(bucket.UploadUrls))
	}
	if bucket.UploadUrls[0].Expiration != times[1] {
		t.Errorf("Expected url[0].Expiration to be times[1], instead got %v", bucket.UploadUrls[0].Expiration)
	}
	if bucket.UploadUrls[1].Expiration != times[3] {
		t.Errorf("Expected url[1].Expiration to be times[3], instead got %v", bucket.UploadUrls[1].Expiration)
	}
}

func createTestFileJson(num int, action Action, fileInfo map[string]string) string {
	file := FileMeta{
		ID:              fmt.Sprintf("id%d", num),
		Name:            fmt.Sprintf("name%d", num),
		Size:            int64(10 + num),
		ContentLength:   int64(10 + num),
		ContentSha1:     "sha1", // TODO make valid SHA1
		ContentType:     "text",
		Action:          action,
		FileInfo:        fileInfo,
		UploadTimestamp: int64(100 + num),
	}
	fileJson, _ := json.Marshal(file)
	return string(fileJson)
}

func createTestUploadUrl() *UploadUrl {
	return &UploadUrl{
		Url:                "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id",
		AuthorizationToken: "token",
		Expiration:         time.Now().UTC().Add(24 * time.Hour),
	}
}
