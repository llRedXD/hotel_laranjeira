package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hotel-laranjeira/backend/internal/auth"
)

func TestReservationStayOperationsRequireAuthentication(t *testing.T) {
	server := &api{auth: auth.NewService(nil, "01234567890123456789012345678901")}
	for _, test := range []struct {
		method, path string
		handler      http.HandlerFunc
	}{
		{http.MethodPost, "/api/v1/reservations/00000000-0000-4000-8000-000000000001/check-in", server.checkInReservation},
		{http.MethodPost, "/api/v1/reservations/00000000-0000-4000-8000-000000000001/check-out", server.checkOutReservation},
		{http.MethodGet, "/api/v1/reservations/00000000-0000-4000-8000-000000000001/check-out-quote?version=2", server.quoteCheckOutReservation},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			test.handler(recorder, httptest.NewRequest(test.method, test.path, nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
