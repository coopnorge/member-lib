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
			expectedError: "",
		},
		{
			name:          "Server Error with Message",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"unexpected_response_detail":"server error"}`,
			expectedError: "",
		},
		{
			name:          "Client Error without Detail",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{}`,
			expectedError: "",
		},
		{
			name:          "Invalid payload json - parsing error",
			statusCode:    http.StatusNetworkAuthenticationRequired,
			responseBody:  `{asd,,,,,unit_test_12341`,
			expectedError: "failed to unmarshal successful response: invalid character 'a' looking for beginning of object key string",
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

			badResponse, parserErr := ExtractErrorResponse(&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
			if tt.responseBody == "" && tt.expectedError == "" {
				assert.Nil(t, badResponse)
				assert.Nil(t, parserErr)
			} else if tt.expectedError == "" {
				assert.NotNil(t, badResponse)

				// NOTE: Check if required data is added to map
				isContaining := false
				for k := range badResponse {
					if assert.Contains(t, tt.responseBody, k) {
						isContaining = !isContaining
					}
				}

				if tt.responseBody != "{}" { // NOTE: Only if not empty Json
					assert.True(t, isContaining, "Must contain needed data in bad response data")
				}
			} else {
				assert.NotNil(t, parserErr)
				assert.True(t, parserErr.Error() == tt.expectedError)
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
	assert.NoError(t, extractorErr, "not expected to have parsing error")
	assert.NotNil(t, extractedBadRequest, "expected to have bad response after request")
}

func TestExtractResponseNoHTTPBOdy(t *testing.T) {
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
	assert.NoError(t, err)
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

	respBody, bodyReadErr := io.ReadAll(resp.Body)
	assert.NoError(t, bodyReadErr)

	_, _, err = ExtractResponse[TestData](&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
	assert.Error(t, err)

	if err != nil && !strings.Contains(err.Error(), "failed to unmarshal successful response") {
		t.Errorf("expected JSON unmarshal error, got %q", err.Error())
	}
}

func TestExtractBadRequestResponse(t *testing.T) {
	type TestData struct {
		Message string `json:"message"`
	}

	exampleOfBadResponse := `{"Detail":"The following error response is returned from SapCrm service: Active card 44444444 exists for COOPID 123456789 Membership 1234567890","Instance":null,"Status":424,"Title":"SapCrmException","Type":"https://tools.ietf.org/html/rfc7231#section-6.5"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFailedDependency)
		fmt.Fprintln(w, exampleOfBadResponse)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	respBody, bodyReadErr := io.ReadAll(resp.Body)
	assert.NoError(t, bodyReadErr)

	okResp, badResponse, parseErr := ExtractResponse[TestData](&Response{HTTPResponse: resp, HTTPResponseBody: &respBody})
	assert.NoError(t, parseErr)
	assert.NotNil(t, badResponse)
	assert.Nil(t, okResp)

	isContainDetails := false
	for k := range badResponse {
		isContainDetails = k == "Detail"
		if isContainDetails {
			break
		}
	}

	assert.True(t, isContainDetails, "expected to be found `Detail` in response body")
}
