package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUntrustedOriginCannotMutate(t *testing.T) {
	called := false
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }), "http://localhost:5173")
	r := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	r.Header.Set("Origin", "https://attacker.example")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if called || w.Code != 403 {
		t.Fatal("untrusted request reached handler")
	}
}

func TestDecodeRejectsUnknownAndTrailingJSON(t *testing.T) {
	for _, body := range []string{`{"unknown":true}`, `{"name":"A"} {"name":"B"}`, `{"name":`} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		var input categoryRequest
		if decodeJSON(w, r, &input) || w.Code != 400 {
			t.Fatalf("accepted %s", body)
		}
	}
}

func TestAuthRateLimit(t *testing.T) {
	handler := security(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for i := 0; i < 31; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/auth/login", nil))
		if i == 30 && w.Code != 429 {
			t.Fatal("rate limit missing")
		}
	}
}

func TestPaginationRejectsUnboundedQueries(t *testing.T) {
	for _, query := range []string{"?limit=1000", "?offset=-1", "?limit=x"} {
		_, _, ok := pagination(httptest.NewRecorder(), httptest.NewRequest("GET", "/"+query, nil))
		if ok {
			t.Fatalf("accepted %s", query)
		}
	}
}
