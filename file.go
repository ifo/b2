package b2

import (
	"fmt"
	"io"
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
	FileMeta
	Data io.ReadWriter
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
	err := b.B2.MakeApiRequest("POST", "/b2api/v1/b2_list_file_names", request, response)
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
	err := b.B2.MakeApiRequest("POST", "/b2api/v1/b2_list_file_versions", request, response)
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
	err := b.B2.MakeApiRequest("POST", "/b2api/v1/b2_get_file_info", request, response)
	if err != nil {
		return nil, err
	}
	response.Bucket = b
	return response, nil
}

func (b *Bucket) UploadFile(name string, fileInfo map[string]string, file io.Reader) error {
	return nil
}

func (b *Bucket) GetUploadUrl() (*UploadUrl, error) {
	request := fmt.Sprintf(`{"bucketId":"%s"}`, b.BucketID)
	response := &UploadUrl{Time: time.Now().UTC()}
	err := b.B2.MakeApiRequest("POST", "/b2api/v1/b2_get_upload_url", request, response)
	if err != nil {
		return nil, err
	}
	b.UploadUrls = append(b.UploadUrls, response)
	return response, nil
}

func (b *Bucket) DownloadFileByName(fileName string) (*File, error) {
	return nil, nil
}

func (b *Bucket) DownloadFileByID(fileID string) (*File, error) {
	return nil, nil
}

func (b *Bucket) HideFile(fileName string) error {
	return nil
}

func (b *Bucket) DeleteFileVersion(fileName, fileID string) error {
	return nil
}
