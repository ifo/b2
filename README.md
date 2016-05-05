b2
==

b2 is a [Backblaze B2 Cloud Storage]
(https://www.backblaze.com/b2/cloud-storage.html) API client written in Go.

## Usage

To use b2, import it and create an API client using your account ID and
application key.
```go
import (
	"github.com/ifo/b2"
)

b2api, err := b2.CreateB2(accountID, appKey)
```

#### Here are some examples:

Create a bucket:
```go
bucket, err := b2api.CreateBucket("kitten-pictures", AllPublic)
```

Upload a file:
```go
fileReader, err := os.Open("./path/to/kitten.jpg")
// handle err

fileMeta, err := bucket.UploadFile("kitten.jpg", fileReader, nil)
```

Download a file:
```go
kittenFile, err := bucket.DownloadFileByName("kitten.jpg")
// handle err

err = ioutil.WriteFile(kittenFile.Meta.Name, kittenFile.Data, 0644)
```

Check for an API error:
```go
kittenFile, err := bucket.DownloadFileByName("cat.jpg")
if err != nil {
	if err, ok := err.(APIError); ok {
		// this is an APIError
		fmt.Println(err.Message)
	}
}
```

## TODO

- Implement large file API
- Integration tests
- Example program

## License

b2 is ISC licensed.
Check out the [LICENSE](https://github.com/ifo/b2/blob/master/LICENSE) file.
