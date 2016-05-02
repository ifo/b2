package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestBucket_ListfileNames(t *testing.T) {
	bucket := testBucket()
	bucket.ListFileNames("name", 1)
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_ListFileVersions(t *testing.T) {
	bucket := testBucket()
	resp, err := bucket.ListFileVersions("", "id", 1)
	if err == nil {
		t.Error("Expected err to exist")
	}
	if resp != nil {
		t.Errorf("Expected resp to be nil, instead got %+v", resp)
	}

	bucket.ListFileVersions("name", "id", 1)
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_parseListFile(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}
	setupFiles := ""
	for i := range fileAction {
		setupFiles += testFileJSON(i, fileAction[i], nil)
		setupFiles += ","
	}
	setupFiles = setupFiles[:len(setupFiles)-1]
	resp := testResponse(200, fmt.Sprintf(`{"files":[%s],"nextFileId":"id%d","nextFileName":"name%d"}`,
		setupFiles, len(fileAction), len(fileAction)))

	bucket := testBucket()
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

	resps := testResponseErrors()
	for i, resp := range resps {
		fileList, err := bucket.parseListFile(resp)
		checkResponseError(err, 400+i, t)
		if fileList != nil {
			t.Errorf("Expected fileList to be empty, instead got %+v", fileList)
		}
	}
}

func TestBucket_GetFileInfo(t *testing.T) {
	bucket := testBucket()
	resp, err := bucket.GetFileInfo("")
	if err.Error() != "No fileID provided" {
		t.Errorf(`Expected "No fileID provided", instead got %s`, err)
	}
	if resp != nil {
		t.Errorf("Expected resp to be nil, instead got %+v", resp)
	}

	bucket.GetFileInfo("id")
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_UploadFile(t *testing.T) {
	bucket := testBucket()
	resp, err := bucket.UploadFile("", nil, nil)
	if err == nil {
		t.Error("Expected err to exist")
	}
	if resp != nil {
		t.Errorf("Expected resp to be nil, instead got %+v", resp)
	}

	resp, err = bucket.UploadFile("name", nil, nil)
	if err == nil {
		t.Error("Expected err to exist")
	}
	if resp != nil {
		t.Errorf("Expected resp to be nil, instead got %+v", resp)
	}

	bucket.UploadURLs = []*UploadURL{testUploadURL()}
	bucket.UploadFile("name", bytes.NewReader([]byte("cats")), nil)
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_setupUploadFile(t *testing.T) {
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

	uploadURLs := []*UploadURL{
		{URL: "https://example.com/1", AuthorizationToken: "token1", Expiration: time.Now().UTC()}, // expired
		{URL: "https://example.com/2", AuthorizationToken: "token2", Expiration: time.Now().UTC().Add(1 * time.Hour)},
	}
	bucket := testBucket()
	bucket.UploadURLs = uploadURLs
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

func TestBucket_GetUploadURL(t *testing.T) {
	bucket := testBucket()
	bucket.GetUploadURL()
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_parseGetUploadURL(t *testing.T) {
	uploadURLStr := "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id"
	resp := testResponse(200, fmt.Sprintf(`{"bucketId":"id","uploadUrl":"%s","authorizationToken":"token"}`, uploadURLStr))

	bucket := testBucket()
	uploadURL, err := bucket.parseGetUploadURL(resp)
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if uploadURL.Expiration.IsZero() {
		t.Error("Expected time to be now + 24h, instead got zero time")
	}
	if uploadURL.AuthorizationToken != "token" {
		t.Errorf(`Expected uploadURL token to be "token", instead got %s`, uploadURL.AuthorizationToken)
	}
	if uploadURL.URL != uploadURLStr {
		t.Errorf("Expected uploadURL's url to be uploadURLStr, instead got %s", uploadURL.URL)
	}

	if len(bucket.UploadURLs) != 1 {
		t.Fatalf("Expected length of bucket upload urls to be 1, insetad was %d", len(bucket.UploadURLs))
	}
	if bucket.UploadURLs[0] != uploadURL {
		t.Error("Expected bucket's first uploadURL to be uploadURL, instead was", bucket.UploadURLs[0])
	}

	resps := testResponseErrors()
	for i, resp := range resps {
		bucket := testBucket()
		uploadURL, err := bucket.parseGetUploadURL(resp)
		checkResponseError(err, 400+i, t)
		if uploadURL != nil {
			t.Errorf("Expected response to be empty, instead got %+v", uploadURL)
		}
	}
}

func TestBucket_DownloadFileByName(t *testing.T) {
	bucket := testBucket()
	bucket.DownloadFileByName("name")
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}

	// public buckets don't need authorization
	bucket.BucketType = AllPublic
	bucket.DownloadFileByName("name")
	req = bucket.B2.client.(*dummyClient).Req
	auth, ok = req.Header["Authorization"]
	if ok {
		t.Errorf("Expected auth to be empty, instead got %s", auth)
	}
}

func TestBucket_DownloadFileByID(t *testing.T) {
	bucket := testBucket()
	bucket.DownloadFileByID("id")
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}

	// public buckets don't need authorization
	bucket.BucketType = AllPublic
	bucket.DownloadFileByID("id")
	req = bucket.B2.client.(*dummyClient).Req
	auth, ok = req.Header["Authorization"]
	if ok {
		t.Errorf("Expected auth to be empty, instead got %s", auth)
	}
}

func TestBucket_parseFile(t *testing.T) {
	headers := map[string][]string{
		"X-Bz-File-Id":      {"1"},
		"X-Bz-File-Name":    {"cats.txt"},
		"Content-Length":    {"19"},
		"X-Bz-Content-Sha1": {"78498e5096b20e3f1c063e8740ff83d595ededb3"},
		"Content-Type":      {"text/plain"},
	}
	fileData := "cats cats cats cats"
	resp := testResponse(200, fileData)
	resp.Header = headers

	bucket := testBucket()
	file, err := bucket.parseFile(resp)
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

	resps := testResponseErrors()
	for i, resp := range resps {
		bucket := testBucket()
		uploadURL, err := bucket.parseFile(resp)
		checkResponseError(err, 400+i, t)
		if uploadURL != nil {
			t.Errorf("Expected response to be empty, instead got %+v", uploadURL)
		}
	}
}

func TestBucket_HideFile(t *testing.T) {
	bucket := testBucket()
	bucket.HideFile("name")
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_DeleteFileVersion(t *testing.T) {
	bucket := testBucket()
	bucket.DeleteFileVersion("name", "id")
	req := bucket.B2.client.(*dummyClient).Req
	auth, ok := req.Header["Authorization"]
	if !ok || auth[0] != bucket.B2.AuthorizationToken {
		t.Errorf("Expected auth to be %s, instead got %s", bucket.B2.AuthorizationToken, auth)
	}
}

func TestBucket_parseFileMeta(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}

	for i := range fileAction {
		resp := testResponse(200, testFileJSON(i, fileAction[i], nil))

		bucket := testBucket()

		fileMeta, err := bucket.parseFileMeta(resp)
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

	resps := testResponseErrors()
	for i, resp := range resps {
		bucket := testBucket()
		fileMeta, err := bucket.parseFileMeta(resp)
		checkResponseError(err, 400+i, t)
		if fileMeta != nil {
			t.Errorf("Expected response to be empty, instead got %+v", fileMeta)
		}
	}
}
func TestBucket_cleanUploadURLs(t *testing.T) {
	bucket := testBucket()

	times := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(1 * time.Hour),
		time.Now().UTC().Add(-1 * time.Hour),
		time.Now().UTC().Add(2 * time.Hour),
	}
	// two UploadURLs should be cleaned
	bucket.UploadURLs = append(bucket.UploadURLs, &UploadURL{Expiration: times[0]})
	bucket.UploadURLs = append(bucket.UploadURLs, &UploadURL{Expiration: times[1]})
	bucket.UploadURLs = append(bucket.UploadURLs, &UploadURL{Expiration: times[2]})
	bucket.UploadURLs = append(bucket.UploadURLs, &UploadURL{Expiration: times[3]})

	bucket.cleanUploadURLs()

	if len(bucket.UploadURLs) != 2 {
		t.Fatalf("Expected UploadURLs length to be 2, instead got %d", len(bucket.UploadURLs))
	}
	if bucket.UploadURLs[0].Expiration != times[1] {
		t.Errorf("Expected url[0].Expiration to be times[1], instead got %v", bucket.UploadURLs[0].Expiration)
	}
	if bucket.UploadURLs[1].Expiration != times[3] {
		t.Errorf("Expected url[1].Expiration to be times[3], instead got %v", bucket.UploadURLs[1].Expiration)
	}
}

func testFileJSON(num int, action Action, fileInfo map[string]string) string {
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
	fileJSON, _ := json.Marshal(file)
	return string(fileJSON)
}

func testUploadURL() *UploadURL {
	return &UploadURL{
		URL:                "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id",
		AuthorizationToken: "token",
		Expiration:         time.Now().UTC().Add(24 * time.Hour),
	}
}
