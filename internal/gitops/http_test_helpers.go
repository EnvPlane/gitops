package gitops

import (
	"net/http"
	"net/http/httptest"
)

type roundTripHandler struct {
	handler http.Handler
}

func (h roundTripHandler) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	h.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}

func newHTTPClientForTest(handler http.Handler) *http.Client {
	return &http.Client{
		Transport: roundTripHandler{handler: handler},
	}
}
