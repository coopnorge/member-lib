package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseNil(t *testing.T) {
	extractedErrorResponse, extractorErr := ExtractErrorResponse(nil)

	assert.Nil(t, extractedErrorResponse)
	assert.NotNil(t, extractorErr)
	assert.True(t, extractorErr.Error() == "http error: response not exist")
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
			expectedError: "http response contains not successful response status (400 - Bad Request) no payload details",
		},
		{
			name:          "Server Error with Message",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"unexpected_response_detail":"server error"}`,
			expectedError: "http response contains not successful response status (500 - Internal Server Error) no payload details",
		},
		{
			name:          "Client Error without Detail",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{}`,
			expectedError: "http response contains not successful response status (400 - Bad Request) no payload details",
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

			respBody, bodyReadErr := io.ReadAll(resp.Body)
			assert.NoError(t, bodyReadErr)

			errResp, extractErr := ExtractErrorResponse(&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
			if (tt.expectedError == "" && extractErr != nil) || (tt.expectedError != "" && (extractErr == nil || extractErr.Error() != tt.expectedError)) {
				t.Errorf("expected error %q, got %q", tt.expectedError, extractErr)
			}

			if tt.expectedError != "" {
				assert.NotNil(t, errResp)
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

	respBody, bodyReadErr := io.ReadAll(resp.Body)
	assert.NoError(t, bodyReadErr)

	ExtractedSuccessfullyResponse, extractedBadResponse, err := ExtractResponse[TestData](&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
	assert.Nil(t, extractedBadResponse)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ExtractedSuccessfullyResponse.Message != testData.Message {
		t.Errorf("expected message %q, got %q", testData.Message, ExtractedSuccessfullyResponse.Message)
	}
}

func TestExtractResponseWithError(t *testing.T) {
	type TestData struct {
		Message string `json:"message"`
	}
	type TestBadData struct {
		Detail string `json:"detail,omitempty"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Set the error status
		_ = json.NewEncoder(w).Encode(TestBadData{Detail: "invalid request"})
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	respBody, bodyReadErr := io.ReadAll(resp.Body)
	assert.NoError(t, bodyReadErr)

	_, extractedBadRequest, extractorErr := ExtractResponse[TestData](&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
	if extractorErr == nil {
		t.Errorf("expected an error, got nil")
	}

	expectedError := "http response contains not successful response status (400 - Bad Request) no payload details"
	if extractorErr != nil && extractorErr.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, extractorErr.Error())
	}

	assert.NotNil(t, extractedBadRequest)
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

	_, _, err = ExtractResponse[TestData](&Response{HTTPResponse: resp})
	if err == nil {
		t.Errorf("expected an error, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "failed to unmarshal successful response") {
		t.Errorf("expected JSON unmarshal error, got %q", err.Error())
	}
}
