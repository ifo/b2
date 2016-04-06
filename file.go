package b2

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type FileMeta struct {
	ID              string            `json:"fileId"`
	Name            string            `json:"fileName"`
	Size            int64             `json:"size"`
	ContentLength   int64             `json:"contentLength"`
	ContentSha1     string            `json:"contentSha1"`
	ContentType     string            `json:"contentType"`
	Action          Action            `json:"action"`
	FileInfo        map[string]string `json:"fileInfo"`
	UploadTimestamp int64             `json:"uploadTimestamp"`
	Bucket          *Bucket           `json:"-"`
}

type Action string

const (
	ActionUpload Action = "upload"
	ActionHide   Action = "hide"
	ActionStart  Action = "start"
)

type File struct {
	Meta FileMeta
	Data []byte
}

type listFileRequest struct {
	BucketID      string `json:"bucketId"`
	StartFileName string `json:"startFileName,omitempty"`
	StartFileID   string `json:"startFileId,omitempty"`
	MaxFileCount  int64  `json:"maxFileCount,omitempty"`
}

type ListFileResponse struct {
	Files        []FileMeta `json:"files"`
	NextFileName string     `json:"nextFileName"`
	NextFileID   string     `json:"nextFileId"`
}

type fileMetaRequest struct {
	BucketID string `json:"bucketId,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileID   string `json:"fileId,omitempty"`
}

func (b *Bucket) ListFileNames(startFileName string, maxFileCount int64) (*ListFileResponse, error) {
	requestBody := listFileRequest{
		StartFileName: startFileName,
		MaxFileCount:  maxFileCount,
	}
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_list_file_names", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseListFile(resp)
}

func (b *Bucket) ListFileVersions(startFileName, startFileID string, maxFileCount int64) (*ListFileResponse, error) {
	if startFileID != "" && startFileName == "" {
		return nil, fmt.Errorf("If startFileID is provided, startFileName must be provided")
	}

	requestBody := listFileRequest{
		BucketID:      b.BucketID,
		StartFileName: startFileName,
		StartFileID:   startFileID,
		MaxFileCount:  maxFileCount,
	}
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_list_file_names", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseListFile(resp)
}

func (b *Bucket) parseListFile(resp *http.Response) (*ListFileResponse, error) {
	defer resp.Body.Close()
	respBody := &ListFileResponse{}
	err := ParseResponse(resp, respBody)
	if err != nil {
		return nil, err
	}

	for i := range respBody.Files {
		respBody.Files[i].Bucket = b
	}
	return respBody, nil
}

func (b *Bucket) GetFileInfo(fileID string) (*FileMeta, error) {
	if fileID == "" {
		return nil, fmt.Errorf("No fileID provided")
	}
	requestBody := fileMetaRequest{FileID: fileID}
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_get_file_info", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMetaResponse(resp)
}

func (b *Bucket) UploadFile(name string, file io.Reader, fileInfo map[string]string) (*FileMeta, error) {
	b.cleanUploadUrls()

	uploadUrl := &UploadUrl{}
	var err error
	if len(b.UploadUrls) > 0 {
		// TODO don't just pick the first usable url
		uploadUrl = b.UploadUrls[0]
	} else {
		uploadUrl, err = b.GetUploadUrl()
		if err != nil {
			return nil, err
		}
	}

	req, err := b.B2.CreateRequest("POST", uploadUrl.Url, file)
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", uploadUrl.AuthorizationToken)
	req.Header.Set("X-Bz-File-Name", "")
	req.Header.Set("Content-Type", "b2/x-auto") // TODO include type if known
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	req.Header.Set("X-Bz-Content-Sha1", fmt.Sprintf("%x", sha1.Sum(fileBytes)))
	for k, v := range fileInfo {
		req.Header.Set("X-Bz-Info-"+k, v)
	}
	// TODO include X-Bz-Info-src_last_modified_millis

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMetaResponse(resp)
}

func (b *Bucket) GetUploadUrl() (*UploadUrl, error) {
	requestBody := fmt.Sprintf(`{"bucketId":"%s"}`, b.BucketID)
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_get_upload_url", requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseGetUploadUrlResponse(resp)
}

func (b *Bucket) parseGetUploadUrlResponse(resp *http.Response) (*UploadUrl, error) {
	defer resp.Body.Close()

	uploadUrl := &UploadUrl{Expiration: time.Now().UTC().Add(24 * time.Hour)}
	err := ParseResponse(resp, uploadUrl)
	if err != nil {
		return nil, err
	}
	b.UploadUrls = append(b.UploadUrls, uploadUrl)
	return uploadUrl, nil
}

func (b *Bucket) DownloadFileByName(fileName string) (*File, error) {
	req, err := b.B2.CreateRequest("GET", b.B2.DownloadUrl+"/file/"+fileName, nil)
	if err != nil {
		return nil, err
	}

	if b.BucketType == AllPrivate {
		req.Header.Set("Authorization", b.B2.AuthorizationToken)
	}

	// ignoring the "Range" header
	// that will be in the file part section (when added)

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileResponse(resp)
}

func (b *Bucket) DownloadFileByID(fileID string) (*File, error) {
	req, err := b.B2.CreateRequest("GET", b.B2.DownloadUrl+"/b2api/v1/b2_download_file_by_id?fileId="+fileID, nil)
	if err != nil {
		return nil, err
	}

	if b.BucketType == AllPrivate {
		req.Header.Set("Authorization", b.B2.AuthorizationToken)
	}

	// ignoring the "Range" header
	// that will be in the file part section (when added)

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileResponse(resp)
}

func (b *Bucket) parseFileResponse(resp *http.Response) (*File, error) {
	defer resp.Body.Close()

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		errJson := errorResponse{}
		if err := json.Unmarshal(fileBytes, &errJson); err != nil {
			return nil, err
		}

		return nil, errJson
	}

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}

	if fmt.Sprintf("%x", sha1.Sum(fileBytes)) != resp.Header.Get("X-Bz-Content-Sha1") {
		// TODO? retry download
		return nil, fmt.Errorf("File sha1 didn't match provided sha1")
	}

	// TODO collect "X-Bz-Info-*" headers

	return &File{
		Meta: FileMeta{
			ID:            resp.Header.Get("X-Bz-File-Id"),
			Name:          resp.Header.Get("X-Bz-File-Name"),
			Size:          int64(len(fileBytes)),
			ContentLength: int64(contentLength),
			ContentSha1:   resp.Header.Get("X-Bz-Content-Sha1"),
			ContentType:   resp.Header.Get("Content-Type"),
			FileInfo:      nil,
			Bucket:        b,
		},
		Data: fileBytes,
	}, nil
}

func (b *Bucket) HideFile(fileName string) (*FileMeta, error) {
	requestBody := fileMetaRequest{
		BucketID: b.BucketID,
		FileName: fileName,
	}
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_hide_file", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMetaResponse(resp)
}

func (b *Bucket) DeleteFileVersion(fileName, fileID string) (*FileMeta, error) {
	requestBody := fileMetaRequest{
		FileName: fileName,
		FileID:   fileID,
	}
	req, err := b.B2.CreateRequest("POST", b.B2.ApiUrl+"/b2api/v1/b2_delete_file_version", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMetaResponse(resp)
}

func (b *Bucket) parseFileMetaResponse(resp *http.Response) (*FileMeta, error) {
	defer resp.Body.Close()
	respBody := &FileMeta{}
	err := ParseResponse(resp, respBody)
	if err != nil {
		return nil, err
	}

	respBody.Bucket = b
	return respBody, nil
}

func (b *Bucket) cleanUploadUrls() {
	if len(b.UploadUrls) == 0 {
		return
	}

	now := time.Now().UTC()
	remainingUrls := []*UploadUrl{}
	for _, url := range b.UploadUrls {
		if url.Expiration.After(now) {
			remainingUrls = append(remainingUrls, url)
		}
	}
	b.UploadUrls = remainingUrls
}
