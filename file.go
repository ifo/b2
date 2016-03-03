package b2

import (
	"io"
)

type FileMeta struct {
	ID            string
	Name          string
	ContentLength int64
	ContentSha1   string
	ContentType   string
	FileInfo      map[string]string
	AccountID     string
	BucketID      string
}

type File struct {
	FileMeta
	Data io.ReadWriter
}

func (b *Bucket) ListFileNames(startFileName string,
	maxFileCount int64) ([]FileMeta, error) {
	return nil, nil
}

func (b *Bucket) ListFileVersions(startFileName, startFileID string,
	maxFileCount int64) ([]FileMeta, error) {
	return nil, nil
}

func (b *Bucket) DownloadFileByID(fileID string) (*File, error) {
	return nil, nil
}

func (b *Bucket) DownloadFileByName(fileName string) (*File, error) {
	return nil, nil
}

func (b *Bucket) UploadFile(name string, fileInfo map[string]string,
	file io.Reader) error {
	return nil
}

func (b *Bucket) GetFileInfo(fileID string) (*FileMeta, error) {
	return nil, nil
}

func (b *Bucket) HideFile(fileName string) error {
	return nil
}

func (b *Bucket) DeleteFileVersion(fileName, fileID string) error {
	return nil
}
