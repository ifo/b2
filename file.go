package b2

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// FileMeta is the meta information of a File, not including the file data.
// It contains a reference to the bucket that the file is within.
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

// Action is the state of a file.
type Action string

// Files may be in the upload (complete), hide, or start (incomplete) state.
const (
	ActionUpload Action = "upload"
	ActionHide   Action = "hide"
	ActionStart  Action = "start"
)

// File is the meta information of a file with its corresponding data.
type File struct {
	Meta FileMeta
	Data []byte
}

// listFileRequest is used for listing the contents of a bucket.
type listFileRequest struct {
	BucketID      string `json:"bucketId"`
	StartFileName string `json:"startFileName,omitempty"`
	StartFileID   string `json:"startFileId,omitempty"`
	MaxFileCount  int64  `json:"maxFileCount,omitempty"`
}

// ListFileResponse is a list of files in a bucket and information regarding
// any files remaining in the bucket that have not been listed.
type ListFileResponse struct {
	Files        []FileMeta `json:"files"`
	NextFileName string     `json:"nextFileName"`
	NextFileID   string     `json:"nextFileId"`
}

// fileMetaRequest is used for any request which returns FileMeta.
type fileMetaRequest struct {
	BucketID string `json:"bucketId,omitempty"`
	FileName string `json:"fileName,omitempty"`
	FileID   string `json:"fileId,omitempty"`
}

// ListFileNames returns FileMeta information for maxCount number of files
// in a bucket, starting with the startName file.
//
// The returned ListFileResponse includes the next name and next ID of
// a file, which can be used to call ListFileNames again.
func (b *Bucket) ListFileNames(startName string, maxCount int64) (*ListFileResponse, error) {
	lfr := listFileRequest{
		BucketID:      b.ID,
		StartFileName: startName,
		MaxFileCount:  maxCount,
	}
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_list_file_names", lfr)
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

// ListFileVersions returns FileMeta on different versions of files.
//
// If a starting file ID is provided, a starting file name must also be given.
func (b *Bucket) ListFileVersions(startName, startID string, maxCount int64) (*ListFileResponse, error) {
	if startID != "" && startName == "" {
		return nil, fmt.Errorf("If startID is provided, startName must be provided")
	}

	lfr := listFileRequest{
		BucketID:      b.ID,
		StartFileName: startName,
		StartFileID:   startID,
		MaxFileCount:  maxCount,
	}
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_list_file_names", lfr)
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

// parseListFile returns a list of FileMeta information
// and a next file name and next file ID.
func (b *Bucket) parseListFile(resp *http.Response) (*ListFileResponse, error) {
	lfr := &ListFileResponse{}
	err := parseResponse(resp, lfr)
	if err != nil {
		return nil, err
	}

	for i := range lfr.Files {
		lfr.Files[i].Bucket = b
	}
	return lfr, nil
}

// GetFileInfo returns FileMeta for the provided fileID.
func (b *Bucket) GetFileInfo(fileID string) (*FileMeta, error) {
	if fileID == "" {
		return nil, fmt.Errorf("No fileID provided")
	}
	fmr := fileMetaRequest{FileID: fileID}
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_get_file_info", fmr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMeta(resp)
}

// UploadFile uploads a file to B2, returning its associated FileMeta info.
//
// The sha1 hash of the file is calculated and included in the upload info.
// If the bucket does not have an UploadURL, one is requested and used.
func (b *Bucket) UploadFile(name string, file io.Reader, fileInfo map[string]string) (*FileMeta, error) {
	if name == "" {
		return nil, fmt.Errorf("No file name provided")
	}
	if file == nil {
		return nil, fmt.Errorf("No file data provided")
	}
	if len(fileInfo) > 10 {
		return nil, fmt.Errorf("More than 10 file info keys provided")
	}
	req, err := b.setupUploadFile(name, file, fileInfo)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMeta(resp)
}

// setupUploadFile sets the required request headers, calculates the sha1 hash,
// and retrieves a new UploadURL if necessary.
//
// It removes all expired UploadURLs before determining if a new one is needed.
// It returns the constructed *http.Request.
func (b *Bucket) setupUploadFile(name string, file io.Reader, fileInfo map[string]string) (*http.Request, error) {
	b.cleanUploadURLs()

	url := &UploadURL{}
	var err error
	if len(b.UploadURLs) > 0 {
		// TODO don't just pick the first usable url
		url = b.UploadURLs[0]
	} else {
		url, err = b.GetUploadURL()
		if err != nil {
			return nil, err
		}
	}

	req, err := CreateRequest("POST", url.URL, file)
	if err != nil {
		return nil, err
	}

	bts, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// TODO percent-encode header values
	req.Header.Set("Authorization", url.AuthorizationToken)
	req.Header.Set("X-Bz-File-Name", name)
	req.Header.Set("Content-Type", "b2/x-auto") // TODO include type if known
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bts)))
	req.Header.Set("X-Bz-Content-Sha1", fmt.Sprintf("%x", sha1.Sum(bts)))
	for k, v := range fileInfo {
		req.Header.Set("X-Bz-Info-"+k, v)
	}
	// TODO include X-Bz-Info-src_last_modified_millis
	// TODO check for total headers being greater than 7,000 bytes

	return req, nil
}

// GetUploadURL gets an UploadURL and attaches it to the associated Bucket.
//
// A new UploadURL will be obtained even if a valid one already exists.
func (b *Bucket) GetUploadURL() (*UploadURL, error) {
	body := fmt.Sprintf(`{"bucketId":"%s"}`, b.ID)
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_get_upload_url", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseGetUploadURL(resp)
}

// parseGetUploadURL parses the GetUploadURL response and attaches the new
// UploadURL to the bucket if successful.
func (b *Bucket) parseGetUploadURL(resp *http.Response) (*UploadURL, error) {
	url := &UploadURL{Expiration: time.Now().UTC().Add(24 * time.Hour)}
	err := parseResponse(resp, url)
	if err != nil {
		return nil, err
	}
	b.UploadURLs = append(b.UploadURLs, url)
	return url, nil
}

// DownloadFileByName gets a File from B2 given the file's name.
//
// If the Bucket is private, Authorization will be set automatically.
func (b *Bucket) DownloadFileByName(name string) (*File, error) {
	req, err := CreateRequest("GET", b.B2.DownloadURL+"/file/"+name, nil)
	if err != nil {
		return nil, err
	}

	if b.Type == AllPrivate {
		req.Header.Set("Authorization", b.B2.AuthorizationToken)
	}

	// ignoring the "Range" header
	// that will be in the file part section (when added)

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFile(resp)
}

// DownloadFileByID gets a File from B2 given the file's ID.
//
// If the Bucket is private, Authorization will be set automatically.
func (b *Bucket) DownloadFileByID(id string) (*File, error) {
	req, err := CreateRequest("GET", b.B2.DownloadURL+"/b2api/v1/b2_download_file_by_id?fileId="+id, nil)
	if err != nil {
		return nil, err
	}

	if b.Type == AllPrivate {
		req.Header.Set("Authorization", b.B2.AuthorizationToken)
	}

	// ignoring the "Range" header
	// that will be in the file part section (when added)

	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFile(resp)
}

// parseFile turns a download file response into a *File.
func (b *Bucket) parseFile(resp *http.Response) (*File, error) {
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parseAPIError(resp)
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	clen, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}

	if fmt.Sprintf("%x", sha1.Sum(bts)) != resp.Header.Get("X-Bz-Content-Sha1") {
		// TODO? retry download
		return nil, fmt.Errorf("File sha1 didn't match provided sha1")
	}

	return &File{
		Meta: FileMeta{
			ID:            resp.Header.Get("X-Bz-File-Id"),
			Name:          resp.Header.Get("X-Bz-File-Name"),
			Size:          int64(len(bts)),
			ContentLength: int64(clen),
			ContentSha1:   resp.Header.Get("X-Bz-Content-Sha1"),
			ContentType:   resp.Header.Get("Content-Type"),
			FileInfo:      GetBzInfoHeaders(resp),
			Bucket:        b,
		},
		Data: bts,
	}, nil
}

// HideFile prevents a named file from being returned during a ListFileNames
// or DownloadFileByName request.
//
// Hidden files may still be found by ID.
func (b *Bucket) HideFile(name string) (*FileMeta, error) {
	fmr := fileMetaRequest{
		BucketID: b.ID,
		FileName: name,
	}
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_hide_file", fmr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMeta(resp)
}

// DeleteFileVersion removes a file version from B2.
//
// If older versions of the same file exist, getting the file by name will
// return the newest of the old versions.
func (b *Bucket) DeleteFileVersion(fileName, fileID string) (*FileMeta, error) {
	// TODO error when fileName or fileID are ""
	fmr := fileMetaRequest{
		FileName: fileName,
		FileID:   fileID,
	}
	req, err := CreateRequest("POST", b.B2.APIURL+"/b2api/v1/b2_delete_file_version", fmr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.B2.AuthorizationToken)
	resp, err := b.B2.client.Do(req)
	if err != nil {
		return nil, err
	}
	return b.parseFileMeta(resp)
}

// parseFileMeta returns the FileMeta of a non-downloading file request.
func (b *Bucket) parseFileMeta(resp *http.Response) (*FileMeta, error) {
	fm := &FileMeta{}
	err := parseResponse(resp, fm)
	if err != nil {
		return nil, err
	}

	fm.Bucket = b
	return fm, nil
}

// cleanUploadURLs deletes UploadURLs that have passed their Expiration date.
func (b *Bucket) cleanUploadURLs() {
	// TODO prevent this from deleting an upload URL that is in use
	// requires upload urls to track self usage
	if len(b.UploadURLs) == 0 {
		return
	}

	now := time.Now().UTC()
	urls := []*UploadURL{}
	for _, url := range b.UploadURLs {
		if url.Expiration.After(now) {
			urls = append(urls, url)
		}
	}
	b.UploadURLs = urls
}
