package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func AssertStatusCode(t *testing.T, response *httptest.ResponseRecorder, expect int) {
	if response.Code != expect {
		t.Errorf("Expect status code %d, but got %d", expect, response.Code)
	}
}

func AssertEqual(t *testing.T, got, want interface{}, message string) {
	if got != want {
		t.Errorf("%s: expect %v, but got %v", message, want, got)
	}
}

func ParseResponse(t *testing.T, data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
}
