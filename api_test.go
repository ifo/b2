package b2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_MakeB2_Success(t *testing.T) {
	s := setupRequest(200, `{"accountId":"1","authorizationToken":"1","apiUrl":"/","downloadUrl":"/"}`)
	defer s.Close()

	b, err := MakeB2("1", "1")
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}

	if b.AccountID != "1" {
		t.Errorf(`Expected AccountID to be "1", instead got %s`, b.AccountID)
	}
	if b.AuthorizationToken != "1" {
		t.Errorf(`Expected AuthorizationToken to be "1", instead got %s`, b.AuthorizationToken)
	}
	if b.ApiUrl != "/" {
		t.Errorf(`Expected AccountID to be "/", instead got %s`, b.ApiUrl)
	}
	if b.DownloadUrl != "/" {
		t.Errorf(`Expected AccountID to be "/", instead got %s`, b.DownloadUrl)
	}
}

func Test_MakeB2_HasAuth(t *testing.T) {
	reqChan := make(chan *http.Request, 1)
	s := setupMockServer(200, "", reqChan)
	defer s.Close()

	MakeB2("1", "1")

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

func Test_MakeB2_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	for i := range codes {
		s := setupRequest(codes[i], bodies[i])

		b, err := MakeB2("1", "1")
		testErrorResponse(err, codes[i], t)
		if b != nil {
			t.Errorf("Expected b to be empty, instead got %+v", b)
		}

		s.Close()
	}
}

func Test_B2_ApiRequest_HasAuth(t *testing.T) {
	reqChan := make(chan *http.Request, 1)
	s := setupMockServer(200, "", reqChan)
	defer s.Close()

	b := makeTestB2()

	b.ApiRequest("GET", "", nil, nil)

	// get the request that the mock server received
	req := <-reqChan

	authToken := req.Header.Get("Authorization")
	if authToken != b.AuthorizationToken {
		t.Errorf("Expected auth token to be %s, instead got %s", b.AuthorizationToken, authToken)
	}
}

func setupRequest(code int, body string) *httptest.Server {
	return setupMockServer(code, body, nil)
}

func setupMockServer(code int, body string, reqChan chan<- *http.Request) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqChan != nil {
			reqChan <- r
		}

		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// set private globals
	httpClient = http.Client{Transport: tr}
	protocol = "http"

	return server
}

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
		t.Fatal("Expected error, no error received")
	}
	if err.Error() !=
		fmt.Sprintf("Status: %d, Code: nope, Message: nope nope", code) {
		t.Errorf(`Expected "Status: %d, Code: nope, Message: nope nope", instead got %s`, code, err)
	}
}

func makeTestB2() *B2 {
	return &B2{
		AccountID:          "id",
		AuthorizationToken: "token",
		ApiUrl:             "https://api900.backblaze.com",
		DownloadUrl:        "https://f900.backblaze.com",
	}
}
