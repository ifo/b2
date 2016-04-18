package b2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_createB2_HasAuth(t *testing.T) {
	reqChan := make(chan *http.Request, 1)
	s, c := setupMockJsonServer(200, "", reqChan)
	defer s.Close()

	client := &client{Protocol: "http", Client: c}

	createB2("1", "1", client)

	// get the request that the mock server received
	req := <-reqChan

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Fatal("Expected ok to be true, instead got false")
	}
	if username != "1" {
		t.Errorf(`Expected username to be "1", instead got %s`, username)
	}
	if password != "1" {
		t.Errorf(`Expected password to be "1", instead got %s`, password)
	}
}

func Test_B2_parseCreateB2Response(t *testing.T) {
	// success
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

	// errors
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
	b2 := &B2{client: &client{Protocol: "https"}} // set client protocol to default

	methods := []string{
		"GET", "POST", "KITTENS",
		"POST", "BAD METHOD",
	}
	urls := []string{
		"http://example.com", "kittens://example.com", "aoeu://example.com",
		"invalid-url", "http://example.com",
	}
	reqBody := struct{ a int }{a: 1}

	for i := 0; i < 3; i++ {
		req, err := b2.CreateRequest(methods[i], urls[i], reqBody)
		if err != nil {
			t.Fatalf("Expected err to be nil, instead got %+v", err)
		}
		if req.URL.Scheme != "https" {
			t.Errorf(`Expected url protocol to be "https", instead got %s`, req.URL.Scheme)
		}
		if req.Body == nil {
			t.Error("Expected req.Body to not be nil")
		}
	}
	for i := 3; i < 5; i++ {
		req, err := b2.CreateRequest(methods[i], urls[i], reqBody)
		if req != nil {
			t.Errorf("Expected req to be nil, instead got %+v", req)
		}
		if err == nil {
			t.Fatal("Expected err to exist")
		}
	}
}

func Test_replaceProtocol(t *testing.T) {
	b2s := []B2{
		{client: &client{Protocol: "https"}},
		{client: &client{Protocol: "http"}},
		{client: &client{Protocol: "kittens"}},
	}
	urls := []string{
		"http://localhost", "https://localhost", "http://localhost", "kittens://localhost",
		"https://www.backblaze.com/", "https://www.backblaze.com/",
		"http://www.backblaze.com/", "kittens://www.backblaze.com/",
		"non/url/",
	}

	for i, b := range b2s {
		index, i2, index2 := i+1, i+4, i+5 // make offsets
		// localhost
		url, err := b.replaceProtocol(urls[i])
		if err != nil {
			t.Errorf("Expected no error, instead got %+v", err)
		}
		if url != urls[index] {
			t.Errorf("Expected url to be %s, instead got %s", urls[index], url)
		}

		// www.backblaze.com
		url, err = b.replaceProtocol(urls[i2])
		if err != nil {
			t.Errorf("Expected no error, instead got %+v", err)
		}
		if url != urls[index2] {
			t.Errorf("Expected url to be %s, instead got %s", urls[index2], url)
		}
	}

	url, err := b2s[0].replaceProtocol(urls[len(urls)-1])
	if err == nil {
		t.Errorf("Expected error, instead got nil, url returned was %s", url)
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

// TODO remove
func setupRequest(code int, body string) (*httptest.Server, http.Client) {
	return setupMockJsonServer(code, body, nil)
}

// TODO remove
func setupMockJsonServer(code int, body string, reqChan chan<- *http.Request) (*httptest.Server, http.Client) {
	headers := map[string]string{"Content-Type": "application/json"}
	return setupMockServer(code, body, headers, reqChan)
}

// TODO remove
func setupMockServer(code int, body string, headers map[string]string, reqChan chan<- *http.Request) (*httptest.Server, http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqChan != nil {
			reqChan <- r
		}

		for k, v := range headers {
			w.Header().Set(k, v)
		}

		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	return server, http.Client{Transport: tr}
}

func createTestResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body))}
}

func createTestErrorResponses() []*http.Response {
	return []*http.Response{
		createTestResponse(400, `{"status":400,"code":"nope","message":"nope nope"}`),
		createTestResponse(401, `{"status":401,"code":"nope","message":"nope nope"}`),
	}
}

// TODO replace with non-200 *http.Response creator
func errorResponses() ([]int, []string) {
	codes := []int{400, 401}
	bodies := []string{
		`{"status":400,"code":"nope","message":"nope nope"}`,
		`{"status":401,"code":"nope","message":"nope nope"}`,
	}
	return codes, bodies
}

func testErrorResponse(err error, code int, t *testing.T) {
	if err == nil {
		t.Error("Expected error, no error received")
	} else if err.Error() != fmt.Sprintf("Status: %d, Code: nope, Message: nope nope", code) {
		t.Errorf(`Expected "Status: %d, Code: nope, Message: nope nope", instead got %s`, code, err)
	}
}

func makeTestB2(c http.Client) *B2 {
	return &B2{
		AccountID:          "id",
		AuthorizationToken: "token",
		ApiUrl:             "https://api900.backblaze.com",
		DownloadUrl:        "https://f900.backblaze.com",
		client:             &client{Protocol: "http", Client: c},
	}
}
