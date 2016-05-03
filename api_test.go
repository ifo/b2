package b2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestB2_createB2(t *testing.T) {
	b2 := testB2()
	b2.createB2()

	req := b2.client.(*testClient).Request
	username, password, ok := req.BasicAuth()
	if !ok {
		t.Fatal("Expected getting Basic Auth to be successful")
	}
	if username != b2.AccountID {
		t.Errorf("Expected username to be %s, instead got %s", b2.AccountID, username)
	}
	if password != b2.ApplicationKey {
		t.Errorf("Expected password to be %s, instead got %s", b2.ApplicationKey, password)
	}
}

func TestB2_parseCreateB2(t *testing.T) {
	resp := testResponse(200, `{"accountId":"1","authorizationToken":"1","apiUrl":"/","downloadUrl":"/"}`)

	b := &B2{AccountID: "1", ApplicationKey: "key"}
	b, err := b.parseCreateB2(resp)
	if err != nil {
		t.Fatalf("Expected err to be nil, instead got %+v", err)
	}

	if b.AccountID != "1" {
		t.Errorf(`Expected AccountID to be "1", instead got %s`, b.AccountID)
	}
	if b.AuthorizationToken != "1" {
		t.Errorf(`Expected AuthorizationToken to be "1", instead got %s`, b.AuthorizationToken)
	}
	if b.APIURL != "/" {
		t.Errorf(`Expected APIURL to be "/", instead got %s`, b.APIURL)
	}
	if b.DownloadURL != "/" {
		t.Errorf(`Expected DownloadURL to be "/", instead got %s`, b.DownloadURL)
	}

	resps := testAPIErrors()
	for i, resp := range resps {
		b := &B2{AccountID: "1", ApplicationKey: "key"}
		b, err := b.parseCreateB2(resp)
		checkAPIError(err, 400+i, t)
		if b != nil {
			t.Errorf("Expected b to be nil, instead got %+v", b)
		}
	}
}

func TestCreateRequest(t *testing.T) {
	methods := []string{"GET", "POST", "BAD METHOD"}
	url := "https://example.com"
	reqBody := struct{ a int }{a: 1}

	i := 0
	for ; i < 2; i++ {
		req, err := CreateRequest(methods[i], url, reqBody)
		if err != nil {
			t.Fatalf("Expected err to be nil, instead got %+v", err)
		}
		if req.Body == nil {
			t.Error("Expected req.Body to not be nil")
		}
		if req.Method != methods[i] {
			t.Errorf("Expected req.Method to be methods[i], instead got %s", req.Method)
		}
		if req.URL.Scheme != "https" {
			t.Errorf(`Expected req.URL.Scheme to be "https", instead got %s`, req.URL.Scheme)
		}
		if req.URL.Host != "example.com" {
			t.Errorf(`Expected req.URL.Host to be "example.com", instead got %s`, req.URL.Host)
		}
	}

	req, err := CreateRequest(methods[i], url, reqBody)
	if err == nil {
		t.Fatal("Expected err to exist")
	}
	if req != nil {
		t.Errorf("Expected req to be nil, instead got %+v", req)
	}
}

func TestGetBzInfoHeaders(t *testing.T) {
	headers := map[string][]string{
		"Content-Type":      {"kittens"},
		"X-Bz-Info-kittens": {"yes"},
		"X-Bz-Info-thing":   {"one"},
	}
	resp := &http.Response{Header: headers}

	bzHeaders := GetBzInfoHeaders(resp)

	if len(bzHeaders) != 2 {
		t.Fatalf("Expected length of headers to be 2, instead got %d", len(bzHeaders))
	}
	if h, ok := bzHeaders["Content-Type"]; ok {
		t.Errorf("Expected no Content-Type, instead recieved %s", h)
	}
	if h := bzHeaders["kittens"]; h != "yes" {
		t.Errorf(`Expected kittens to be "yes", instead got %s`, h)
	}
	if h := bzHeaders["thing"]; h != "one" {
		t.Errorf(`Expected thing to be "one", instead got %s`, h)
	}
}

type testClient struct {
	Request *http.Request
}

func (dc *testClient) Do(r *http.Request) (*http.Response, error) {
	dc.Request = r
	return testResponse(400, `{"status":400,"code":"nope","message":"nope nope"}`), nil
}

func testB2() *B2 {
	return &B2{
		AccountID:          "id",
		AuthorizationToken: "token",
		APIURL:             "https://api900.backblaze.com",
		DownloadURL:        "https://f900.backblaze.com",
		client:             &testClient{},
	}
}

func testResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}

func testAPIErrors() []*http.Response {
	return []*http.Response{
		testResponse(400, `{"status":400,"code":"nope","message":"nope nope"}`),
		testResponse(401, `{"status":401,"code":"nope","message":"nope nope"}`),
	}
}

func checkAPIError(err error, code int, t *testing.T) {
	if err == nil {
		t.Error("Expected error, no error received")
	} else if err.Error() != fmt.Sprintf("Status: %d, Code: nope, Message: nope nope", code) {
		t.Errorf(`Expected "Status: %d, Code: nope, Message: nope nope", instead got %s`, code, err)
	}
}
