package b2

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

func (b *Bucket) ListFileNames(startFileName string, maxFileCount int64) (*ListFileResponse, error) {
	request := listFileRequest{
		BucketID:      b.BucketID,
		StartFileName: startFileName,
		MaxFileCount:  maxFileCount,
	}
	response := &ListFileResponse{}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_list_file_names", request, response)
	if err != nil {
		return nil, err
	}

	for i := range response.Files {
		response.Files[i].Bucket = b
	}

	return response, nil
}

func (b *Bucket) ListFileVersions(startFileName, startFileID string, maxFileCount int64) (*ListFileResponse, error) {
	if startFileID != "" && startFileName == "" {
		return nil, fmt.Errorf("If startFileID is provided, startFileName must be provided")
	}
	request := listFileRequest{
		BucketID:      b.BucketID,
		StartFileName: startFileName,
		StartFileID:   startFileID,
		MaxFileCount:  maxFileCount,
	}
	response := &ListFileResponse{}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_list_file_versions", request, response)
	if err != nil {
		return nil, err
	}

	for i := range response.Files {
		response.Files[i].Bucket = b
	}

	return response, nil
}

func (b *Bucket) GetFileInfo(fileID string) (*FileMeta, error) {
	if fileID == "" {
		return nil, fmt.Errorf("No fileID provided")
	}
	request := fmt.Sprintf(`{"fileId":"%s"}`, fileID)
	response := &FileMeta{}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_get_file_info", request, response)
	if err != nil {
		return nil, err
	}
	response.Bucket = b
	return response, nil
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

	response := &FileMeta{Bucket: b}
	resp, err := httpClientDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = ParseResponseBody(resp, response)
	if err != nil {
		return nil, err
	}
	response.FileInfo = GetBzHeaders(resp)

	return response, nil
}

func (b *Bucket) GetUploadUrl() (*UploadUrl, error) {
	request := fmt.Sprintf(`{"bucketId":"%s"}`, b.BucketID)
	response := &UploadUrl{Expiration: time.Now().UTC().Add(24 * time.Hour)}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_get_upload_url", request, response)
	if err != nil {
		return nil, err
	}
	b.UploadUrls = append(b.UploadUrls, response)
	return response, nil
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

	resp, err := httpClientDo(req)
	if err != nil {
		return nil, err
	}
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

	resp, err := httpClientDo(req)
	if err != nil {
		return nil, err
	}
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
	request := fmt.Sprintf(`{"fileName":"%s","bucketId","%s"}`, fileName, b.BucketID)
	response := &FileMeta{Bucket: b}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_hide_file", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// TODO? return only the fileName and fileId, instead of mostly blank FileMeta
func (b *Bucket) DeleteFileVersion(fileName, fileID string) (*FileMeta, error) {
	request := fmt.Sprintf(`{"fileName":"%s","fileId":"%s"}`, fileName, fileID)
	response := &FileMeta{Bucket: b}
	err := b.B2.ApiRequest("POST", "/b2api/v1/b2_delete_file_version", request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
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
