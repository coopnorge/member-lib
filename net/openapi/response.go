package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response wraps the HTTP response from an API.
type Response struct {
	HTTPResponse     *http.Response // HTTPResponse holds the raw response from the HTTP request.
	HTTPResponseBody *[]byte        // HTTPResponseBody should be the read and stored response body data.
}

// ExtractErrorResponse extracts error details from Response if the HTTP status code indicates an error.
// It returns a map of the error details if an error is present, otherwise nil.
func ExtractErrorResponse(resp *Response) (errResponse map[string]any, extractErr error) {
	if resp == nil {
		return nil, fmt.Errorf("http error: response not exist")
	}

	if resp.HTTPResponse.StatusCode >= http.StatusOK && resp.HTTPResponse.StatusCode < http.StatusMultipleChoices {
		return nil, nil
	}

	if resp.HTTPResponseBody != nil {
		errResponse = map[string]any{}
		if err := json.Unmarshal(*resp.HTTPResponseBody, &errResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal successful response: %w", err)
		}
	}

	extractErr = fmt.Errorf(
		"http response contains not successful response status (%d - %s) no payload details",
		resp.HTTPResponse.StatusCode,
		http.StatusText(resp.HTTPResponse.StatusCode),
	)

	return
}

// ExtractResponse extracts the JSON payload from an Response into T if the HTTP status is successful.
// Any error response placed as map, or an error if the extraction fails.
func ExtractResponse[T any](resp *Response) (successResponse *T, errorResponse map[string]any, extractErr error) {
	errorResponse, extractErr = ExtractErrorResponse(resp)
	if extractErr != nil {
		return nil, errorResponse, extractErr
	}

	if err := json.Unmarshal(*resp.HTTPResponseBody, &successResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal successful response: %w", err)
	}

	return successResponse, nil, nil
}
