package b2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_Bucket_parseListFile(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}
	setupFiles := ""
	for i := range fileAction {
		setupFiles += makeTestFileJson(i, fileAction[i], nil)
		if i != len(fileAction)-1 {
			setupFiles += ","
		}
	}
	resp := createTestResponse(200, fmt.Sprintf(`{"files":[%s],"nextFileId":"id%d","nextFileName":"name%d"}`,
		setupFiles, len(fileAction), len(fileAction)))

	bucket := makeTestBucket(&B2{})
	response, err := bucket.parseListFile(resp)
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

	resps := createTestErrorResponses()
	for i, resp := range resps {
		response, err := bucket.parseListFile(resp)
		testErrorResponse(err, 400+i, t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}
	}
}

func Test_Bucket_ListFileVersions_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		response, err := bucket.ListFileVersions("", "", 0)
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}
}

func Test_Bucket_GetFileInfo_Success(t *testing.T) {
	fileAction := []Action{ActionUpload, ActionHide, ActionStart}

	for i := range fileAction {
		s, c := setupRequest(200, makeTestFileJson(i, fileAction[i], nil))

		b := makeTestB2(c)
		bucket := makeTestBucket(b)

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

	var bucket *Bucket

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket = makeTestBucket(b)
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
	fileInfo := map[string]string{"is_cats": "no :(", "should_be_cats": "of course"}
	respHeaders := map[string]string{}
	for k, v := range fileInfo {
		respHeaders["X-Bz-Info-"+k] = v
	}

	s, c := setupMockServer(200, makeTestFileJson(0, ActionUpload, fileInfo), respHeaders, nil)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	bucket.UploadUrls = append(bucket.UploadUrls, makeTestUploadUrl())
	fileMeta, err := bucket.UploadFile("name0", strings.NewReader("length ten"), fileInfo)
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
		if val, ok := fileInfo[k]; !ok || val != v {
			t.Errorf(`Expected fileMeta.FileInfo["%s"] to be "%s", instead got "%s"`, k, val, v)
		}
	}
}

func Test_Bucket_UploadFile_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		response, err := bucket.UploadFile("", strings.NewReader(""), nil)
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}

		s.Close()
	}
}

func Test_Bucket_setupUploadFile(t *testing.T) {
	t.Skip()
}

func Test_Bucket_GetUploadUrl_Success(t *testing.T) {

	uploadUrl := "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id"

	s, c := setupRequest(200, fmt.Sprintf(`{"bucketId":"id","uploadUrl":"%s","authorizationToken":"token"}`, uploadUrl))
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
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

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		response, err := bucket.GetUploadUrl()
		testErrorResponse(err, codes[i], t)
		if response != nil {
			t.Errorf("Expected response to be empty, instead got %+v", response)
		}
		if len(bucket.UploadUrls) != 0 {
			t.Errorf("Expected no upload urls, instead got %+v", bucket.UploadUrls)
		}

		s.Close()
	}
}

func Test_Bucket_parseGetUploadUrlResponse(t *testing.T) {
	t.Skip()
}

func Test_Bucket_DownloadFileByName_Success(t *testing.T) {
	headers := map[string]string{
		"X-Bz-File-Id":      "1",
		"X-Bz-File-Name":    "cats.txt",
		"Content-Length":    "20",
		"X-Bz-Content-Sha1": "74c3c58c1b6a2a4f63c187493388b421e732fc4b",
		"Content-Type":      "text/plain",
	}

	fileData := "cats cats cats cats"
	// a newline is added at the end of the bytes
	fileDataBytes := append(bytes.NewBufferString(fileData).Bytes(), 10)

	s, c := setupMockServer(200, fileData, headers, nil)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	file, err := bucket.DownloadFileByName("cats.txt")
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
		t.Errorf("Expected file.Meta.Size to be 20, instead got %d", file.Meta.Size)
	}
	if file.Meta.ContentLength != 20 {
		t.Errorf("Expected file.Meta.ContentLength to be 20, instead got %d", file.Meta.ContentLength)
	}
	if file.Meta.ContentSha1 != headers["X-Bz-Content-Sha1"] {
		t.Errorf(`Expected file.Meta.Sha1 to be "%s", instead got %s`, headers["X-Bz-Content-Sha1"], file.Meta.ContentSha1)
	}
	if file.Meta.ContentType != "text/plain" {
		t.Errorf(`Expected file.Meta.ContentType to be "text/plain", instead got %s`, file.Meta.ContentType)
	}
	// TODO include and test fileinfo
	for k, v := range file.Meta.FileInfo {
		t.Errorf("Expected fileInfo to be blank, instead got %s, %s", k, v)
	}
	if !bytes.Equal(file.Data, fileDataBytes) {
		t.Errorf(`Expected file.Data to be "%v", instead got %v`, fileDataBytes, file.Data)
	}

	if file.Meta.Bucket != bucket {
		t.Errorf("Expected file.Meta.bucket to be bucket, instead got %+v", file.Meta.Bucket)
	}
}

func Test_Bucket_DownloadFileByName_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		file, err := bucket.DownloadFileByName("cat.txt")
		testErrorResponse(err, codes[i], t)
		if file != nil {
			t.Errorf("Expected file to be nil, instead got %+v", file)
		}

		s.Close()
	}
}

func Test_Bucket_DownloadFileByID_Success(t *testing.T) {
	headers := map[string]string{
		"X-Bz-File-Id":      "1",
		"X-Bz-File-Name":    "cats.txt",
		"Content-Length":    "20",
		"X-Bz-Content-Sha1": "74c3c58c1b6a2a4f63c187493388b421e732fc4b",
		"Content-Type":      "text/plain",
	}

	fileData := "cats cats cats cats"
	// a newline is added at the end of the bytes
	fileDataBytes := append(bytes.NewBufferString(fileData).Bytes(), 10)

	s, c := setupMockServer(200, fileData, headers, nil)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	file, err := bucket.DownloadFileByID("1")
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
		t.Errorf("Expected file.Meta.Size to be 20, instead got %d", file.Meta.Size)
	}
	if file.Meta.ContentLength != 20 {
		t.Errorf("Expected file.Meta.ContentLength to be 20, instead got %d", file.Meta.ContentLength)
	}
	if file.Meta.ContentSha1 != headers["X-Bz-Content-Sha1"] {
		t.Errorf(`Expected file.Meta.Sha1 to be "%s", instead got %s`, headers["X-Bz-Content-Sha1"], file.Meta.ContentSha1)
	}
	if file.Meta.ContentType != "text/plain" {
		t.Errorf(`Expected file.Meta.ContentType to be "text/plain", instead got %s`, file.Meta.ContentType)
	}
	// TODO include and test fileinfo
	for k, v := range file.Meta.FileInfo {
		t.Errorf("Expected fileInfo to be blank, instead got %s, %s", k, v)
	}
	if !bytes.Equal(file.Data, fileDataBytes) {
		t.Errorf(`Expected file.Data to be "%v", instead got %v`, fileDataBytes, file.Data)
	}

	if file.Meta.Bucket != bucket {
		t.Errorf("Expected file.Meta.bucket to be bucket, instead got %+v", file.Meta.Bucket)
	}
}

func Test_Bucket_DownloadFileByID_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		file, err := bucket.DownloadFileByID("1")
		testErrorResponse(err, codes[i], t)
		if file != nil {
			t.Errorf("Expected file to be nil, instead got %+v", file)
		}

		s.Close()
	}
}

func Test_Bucket_parseFileResponse(t *testing.T) {
	t.Skip()
}

func Test_Bucket_HideFile_Success(t *testing.T) {
	unixTime := time.Now().Unix()
	s, c := setupRequest(200, fmt.Sprintf(`{
"fileId":"1",
"fileName":"cats.txt",
"contentType":null,
"contentSha1":null,
"fileInfo":{},
"action":"hide",
"size":0,
"uploadTimestamp":%d
}`, unixTime))
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	fileMeta, err := bucket.HideFile("cats.txt")
	if err != nil {
		t.Fatalf("Expected err to be nil, instead got %+v", err)
	}

	if fileMeta.ID != "1" {
		t.Errorf(`Expected fileMeta.ID to be "1", instead got %s`, fileMeta.ID)
	}
	if fileMeta.Name != "cats.txt" {
		t.Errorf(`Expected fileMeta.Name to be "cats.txt", instead got %s`, fileMeta.Name)
	}
	if fileMeta.ContentType != "" {
		t.Errorf(`Expected fileMeta.ContentType to be "", instead got %s`, fileMeta.ContentType)
	}
	if fileMeta.ContentSha1 != "" {
		t.Errorf(`Expected fileMeta.ContentSha1 to be "", instead got %s`, fileMeta.ContentSha1)
	}
	if fileMeta.Action != "hide" {
		t.Errorf(`Expected fileMeta.Action to be "hide", instead got %s`, fileMeta.Action)
	}
	if fileMeta.Size != 0 {
		t.Errorf(`Expected fileMeta.Size to be 0, instead got %d`, fileMeta.Size)
	}
	if fileMeta.UploadTimestamp != unixTime {
		t.Errorf(`Expected fileMeta.UploadTimestamp to be unixTime, instead got %v`, fileMeta.UploadTimestamp)
	}
	for k, v := range fileMeta.FileInfo {
		t.Errorf("Expected fileMeta.FileInfo to be empty, instead got %s: %s", k, v)
	}

	if fileMeta.Bucket != bucket {
		t.Errorf("Expected fileMeta.Bucket to be bucket, instead got %+v", fileMeta.Bucket)
	}
}

func Test_Bucket_HideFile_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		fileMeta, err := bucket.HideFile("cats.txt")
		testErrorResponse(err, codes[i], t)
		if fileMeta != nil {
			t.Errorf("Expected fileMeta to be nil, instead got %+v", fileMeta)
		}

		s.Close()
	}
}

func Test_Bucket_DeleteFileVersion_Success(t *testing.T) {
	s, c := setupRequest(200, `{"fileId":"1","fileName":"cats.txt"}`)
	defer s.Close()

	b := makeTestB2(c)
	bucket := makeTestBucket(b)
	fileMeta, err := bucket.DeleteFileVersion("cats.txt", "1")
	if err != nil {
		t.Fatalf("Expected err to be nil, instead got %+v", err)
	}

	if fileMeta.ID != "1" {
		t.Errorf(`Expected fileMeta.ID to be "1", instead got %s`, fileMeta.ID)
	}
	if fileMeta.Name != "cats.txt" {
		t.Errorf(`Expected fileMeta.Name to be "cats.txt", instead got %s`, fileMeta.Name)
	}

	if fileMeta.Bucket != bucket {
		t.Errorf("Expected fileMeta.Bucket to be bucket, instead got %+v", fileMeta.Bucket)
	}
}

func Test_Bucket_DeleteFileVersion_Errors(t *testing.T) {
	codes, bodies := errorResponses()

	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		b := makeTestB2(c)
		bucket := makeTestBucket(b)
		fileMeta, err := bucket.DeleteFileVersion("cats.txt", "1")
		testErrorResponse(err, codes[i], t)
		if fileMeta != nil {
			t.Errorf("Expected fileMeta to be nil, instead got %+v", fileMeta)
		}

		s.Close()
	}
}

func Test_Bucket_parseFileMetaResponse(t *testing.T) {
	t.Skip()
}

func Test_Bucket_cleanUploadUrls(t *testing.T) {
	b := makeTestB2(http.Client{})
	bucket := makeTestBucket(b)

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

func makeTestFileJson(num int, action Action, fileInfo map[string]string) string {
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

func makeTestUploadUrl() *UploadUrl {
	return &UploadUrl{
		Url:                "https://eg.backblaze.com/b2api/v1/b2_upload_file?cvt=eg&bucket=id",
		AuthorizationToken: "token",
		Expiration:         time.Now().UTC().Add(24 * time.Hour),
	}
}
