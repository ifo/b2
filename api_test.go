package b2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_MakeB2_200(t *testing.T) {
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

func Test_MakeB2_400(t *testing.T) {
	s := setupRequest(400,
		`{"status":400,"code":"nope","message":"nope nope"}`)
	defer s.Close()

	b, err := MakeB2("1", "1")
	if err == nil {
		t.Fatal("Expected error, no error received")
	}
	if err.Error() != "Status: 400, Code: nope, Message: nope nope" {
		t.Errorf(`Expected "Status: 400, Code: nope, Message: nope nope", instead got %s`, err)
	}
	if b != nil {
		t.Errorf("Expected b to be empty, instead got %+v", b)
	}
}

func Test_MakeB2_401(t *testing.T) {
	s := setupRequest(401, `{"status":401,"code":"nope","message":"nope nope"}`)
	defer s.Close()

	b, err := MakeB2("1", "1")
	if err == nil {
		t.Fatal("Expected error, no error received")
	}
	if err.Error() != "Status: 401, Code: nope, Message: nope nope" {
		t.Errorf(`Expected "Status: 401, Code: nope, Message: nope nope", instead got %s`, err)
	}
	if b != nil {
		t.Errorf("Expected b to be empty, instead got %+v", b)
	}
}

func setupRequest(code int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
