package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ResponseProblemDetails generic object that will have details of response.
type ResponseProblemDetails struct {
	Detail string `json:"unexpected_response_detail"`
}

// ResponseError checks if the response contains an HTTP error and returns a descriptive error.
func ResponseError(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("http error: response not exist")
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	var pd ResponseProblemDetails
	if err := json.NewDecoder(resp.Body).Decode(&pd); err == nil {
		if pd.Detail != "" {
			return fmt.Errorf("http error: %d %s", resp.StatusCode, pd.Detail)
		}
	}

	return fmt.Errorf("http error: %d - unable to parse detailed error message", resp.StatusCode)
}

// ExtractResponse is a generic function to extract the successful JSON response.
func ExtractResponse[T any](resp *http.Response) (*T, error) {
	var respData T

	if err := ResponseError(resp); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal successful response: %w", err)
	}

	return &respData, nil
}
