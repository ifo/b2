package b2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func Test_B2_createB2(t *testing.T) {
	b2 := createTestB2()
	b2.createB2()

	req := b2.client.(*dummyClient).Req
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

func Test_B2_parseCreateB2Response(t *testing.T) {
	resp := createTestResponse(200, `{"accountId":"1","authorizationToken":"1","apiUrl":"/","downloadUrl":"/"}`)

	b := &B2{AccountID: "1", ApplicationKey: "key"}
	b, err := b.parseCreateB2Response(resp)
	if err != nil {
		t.Fatalf("Expected err to be nil, instead got %+v", err)
	}

	if b.AccountID != "1" {
		t.Errorf(`Expected AccountID to be "1", instead got %s`, b.AccountID)
	}
	if b.AuthorizationToken != "1" {
		t.Errorf(`Expected AuthorizationToken to be "1", instead got %s`, b.AuthorizationToken)
	}
	if b.ApiUrl != "/" {
		t.Errorf(`Expected ApiUrl to be "/", instead got %s`, b.ApiUrl)
	}
	if b.DownloadUrl != "/" {
		t.Errorf(`Expected DownloadUrl to be "/", instead got %s`, b.DownloadUrl)
	}

	resps := createTestErrorResponses()
	for i, resp := range resps {
		b := &B2{AccountID: "1", ApplicationKey: "key"}
		b, err := b.parseCreateB2Response(resp)
		testErrorResponse(err, 400+i, t)
		if b != nil {
			t.Errorf("Expected b to be nil, instead got %+v", b)
		}
	}
}

func Test_B2_CreateRequest(t *testing.T) {
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
	}

	req, err := CreateRequest(methods[i], url, reqBody)
	if err == nil {
		t.Fatal("Expected err to exist")
	}
	if req != nil {
		t.Errorf("Expected req to be nil, instead got %+v", req)
	}
}

func Test_GetBzInfoHeaders(t *testing.T) {
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

type dummyClient struct {
	Req *http.Request
}

func (dc *dummyClient) Do(req *http.Request) (*http.Response, error) {
	dc.Req = req
	return createTestResponse(400, `{"status":400,"code":"nope","message":"nope nope"}`), nil
}

func createTestB2() *B2 {
	return &B2{
		AccountID:          "id",
		AuthorizationToken: "token",
		ApiUrl:             "https://api900.backblaze.com",
		DownloadUrl:        "https://f900.backblaze.com",
		client:             &dummyClient{},
	}
}

func createTestResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}

func createTestErrorResponses() []*http.Response {
	return []*http.Response{
		createTestResponse(400, `{"status":400,"code":"nope","message":"nope nope"}`),
		createTestResponse(401, `{"status":401,"code":"nope","message":"nope nope"}`),
	}
}

func testErrorResponse(err error, code int, t *testing.T) {
	if err == nil {
		t.Error("Expected error, no error received")
	} else if err.Error() != fmt.Sprintf("Status: %d, Code: nope, Message: nope nope", code) {
		t.Errorf(`Expected "Status: %d, Code: nope, Message: nope nope", instead got %s`, code, err)
	}
}
