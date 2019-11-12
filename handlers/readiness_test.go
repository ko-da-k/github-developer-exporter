package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessHandler(t *testing.T) {
	testHandler := NewReadinessHandler()
	testRecorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/hoge", nil)
	if err != nil {
		t.Fatalf("%+v\n", err)
	}

	testHandler.ServeHTTP(testRecorder, req)

	if status := testRecorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	actual := testRecorder.Body.String()
	if testRecorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body\ngot %v\nwant %v",
			actual, expected)
	}
}
