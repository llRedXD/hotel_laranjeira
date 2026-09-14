package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPatchPresence(t *testing.T) {
	for _, test := range []struct {
		body  string
		set   bool
		value bool
		valid bool
	}{
		{`{}`, false, false, true},
		{`{"email":null}`, true, false, true},
		{`{"email":"a@example.com"}`, true, true, true},
		{`{"email":123}`, false, false, false},
		{`{"unknown":"x"}`, false, false, false},
	} {
		r := httptest.NewRequest("PATCH", "/", strings.NewReader(test.body))
		r.Header.Set("Content-Type", "application/json")
		var p guestPatchRequest
		ok := decodeJSON(httptest.NewRecorder(), r, &p)
		if ok != test.valid {
			t.Fatalf("validade: %s", test.body)
		}
		if ok && (p.Email.Set != test.set || (p.Email.Value != nil) != test.value) {
			t.Fatalf("presença: %s", test.body)
		}
	}
}
