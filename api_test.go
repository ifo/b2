package b2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_createB2_Success(t *testing.T) {
	s, c := setupRequest(200, `{"accountId":"1","authorizationToken":"1","apiUrl":"/","downloadUrl":"/"}`)
	defer s.Close()

	client := &client{Protocol: "http", Client: c}

	b, err := createB2("1", "1", client)
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

func Test_MakeB2_Errors(t *testing.T) {
	codes, bodies := errorResponses()
	for i := range codes {
		s, c := setupRequest(codes[i], bodies[i])

		client := &client{Protocol: "http", Client: c}
		b, err := createB2("1", "1", client)
		testErrorResponse(err, codes[i], t)
		if b != nil {
			t.Errorf("Expected b to be empty, instead got %+v", b)
		}

		s.Close()
	}
}

func Test_parseCreateB2Response(t *testing.T) {
	t.Skip()
}

func setupRequest(code int, body string) (*httptest.Server, http.Client) {
	return setupMockJsonServer(code, body, nil)
}

func setupMockJsonServer(code int, body string, reqChan chan<- *http.Request) (*httptest.Server, http.Client) {
	headers := map[string]string{"Content-Type": "application/json"}
	return setupMockServer(code, body, headers, reqChan)
}

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
