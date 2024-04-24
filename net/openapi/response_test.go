package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseNil(t *testing.T) {
	assert.True(t, ResponseError(nil).Error() == "http error: response not exist")
}

func TestResponseError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "Success Response",
			statusCode:    http.StatusOK,
			responseBody:  "",
			expectedError: "",
		},
		{
			name:          "Client Error with Message",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"unexpected_response_detail":"invalid input"}`,
			expectedError: "http error: 400 invalid input",
		},
		{
			name:          "Server Error with Message",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"unexpected_response_detail":"server error"}`,
			expectedError: "http error: 500 server error",
		},
		{
			name:          "Client Error without Detail",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{}`,
			expectedError: "http error: 400 - unable to parse detailed error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			err = ResponseError(resp)
			if (tt.expectedError == "" && err != nil) || (tt.expectedError != "" && (err == nil || err.Error() != tt.expectedError)) {
				t.Errorf("expected error %q, got %q", tt.expectedError, err)
			}
		})
	}
}

func TestExtractResponse(t *testing.T) {
	type TestData struct {
		Message string `json:"message"`
	}
	testData := TestData{Message: "Success"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(testData)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	receivedData, err := ExtractResponse[TestData](resp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedData.Message != testData.Message {
		t.Errorf("expected message %q, got %q", testData.Message, receivedData.Message)
	}
}

func TestExtractResponseWithError(t *testing.T) {
	type TestData struct {
		Message string `json:"message"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Set the error status
		json.NewEncoder(w).Encode(ResponseProblemDetails{Detail: "invalid request"})
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	_, err = ExtractResponse[TestData](resp)
	if err == nil {
		t.Errorf("expected an error, got nil")
	}

	expectedError := "http error: 400 invalid request"
	if err != nil && err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestExtractResponseJSONDecodeError(t *testing.T) {
	type TestData struct {
		Message string `json:"message"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{bad json")
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	_, err = ExtractResponse[TestData](resp)
	if err == nil {
		t.Errorf("expected an error, got nil")
	}

	if err != nil && !containsSubstring(err.Error(), "failed to unmarshal successful response") {
		t.Errorf("expected JSON unmarshal error, got %q", err.Error())
	}
}

func containsSubstring(fullString, substring string) bool {
	return strings.Contains(fullString, substring)
}
