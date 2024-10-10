package oauth

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// stubClientOAuth implements ClientOAuth interface for testing.
type stubClientOAuth struct{}

func (m *stubClientOAuth) AudiencePayload() ([]byte, error) {
	return json.Marshal(map[string]string{
		"audience": "test-audience",
	})
}

func (m *stubClientOAuth) AccessToken() (Token, error) {
	return Token{}, nil
}

func TestAccessTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		jsonResponse := `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`
		w.Write([]byte(jsonResponse))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		AccessTokenURL: server.URL,
		Transport:      server.Client(),
	}
	client := NewClient(cfg)
	client.ClientOAuth = &stubClientOAuth{}

	token, err := client.AccessToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token.AccessToken != "test-token" {
		t.Errorf("Expected access token to be 'test-token', got '%s'", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type to be 'Bearer', got '%s'", token.TokenType)
	}

	if token.ExpiresIn != 3600 {
		t.Errorf("Expected expires_in to be 3600, got '%d'", token.ExpiresIn)
	}

	// check if client.cachedToken is now not empty and valid
	valid := client.isValidToken()
	assert.True(t, valid)
	assert.NotEmpty(t, client.cachedToken)
}

func TestAccessTokenExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		jsonResponse := `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 1}`
		w.Write([]byte(jsonResponse))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		AccessTokenURL: server.URL,
		Transport:      server.Client(),
	}
	client := NewClient(cfg)
	client.ClientOAuth = &stubClientOAuth{}

	token, err := client.AccessToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token.AccessToken != "test-token" {
		t.Errorf("Expected access token to be 'test-token', got '%s'", token.AccessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type to be 'Bearer', got '%s'", token.TokenType)
	}

	if token.ExpiresIn != 1 {
		t.Errorf("Expected expires_in to be 3600, got '%d'", token.ExpiresIn)
	}

	time.Sleep(time.Second + time.Millisecond*500)

	// check if client.cachedToken is now not empty and valid
	valid := client.isValidToken()
	assert.False(t, valid)

	// Must be give new token
	_, err = client.AccessToken()
	assert.NoError(t, err)

	valid = client.isValidToken()
	assert.True(t, valid)
}

func TestAccessTokenHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &ClientConfig{
		AccessTokenURL: server.URL,
		Transport:      server.Client(),
	}
	client := NewClient(cfg)
	client.ClientOAuth = &stubClientOAuth{}

	_, err := client.AccessToken()
	if err == nil {
		t.Fatal("Expected an error, got none")
	}
}

func TestAccessTokenDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{bad json"))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		AccessTokenURL: server.URL,
		Transport:      server.Client(),
	}
	client := NewClient(cfg)
	client.ClientOAuth = &stubClientOAuth{}

	_, err := client.AccessToken()
	if err == nil {
		t.Fatal("Expected a JSON decode error, got none")
	}
}

// exampleOAuthClient shows how could finally implementation could look like.
// You could create separate struct per API where you might get Token.
type exampleOAuthClient struct {
	*AbstractClient
}

// AudiencePayload implementation it's what you only need to implement for oauth.AbstractClient.
func (c *exampleOAuthClient) AudiencePayload() ([]byte, error) {
	reqBody, marshalErr := json.Marshal(map[string]string{
		"grant_type":    "client_credentials", // depends on server settings
		"client_id":     "your_application_id",
		"client_secret": "your_application_secret",
		"audience":      "your_audience_of_interest",
	})

	if marshalErr != nil {
		return nil, fmt.Errorf("error preparing audience payload: %w", marshalErr)
	}

	return reqBody, nil
}

func Example_newOAuthClient() {
	/*
		// exampleOAuthClient shows how could finally implementation could look like.
		// You could create separate struct per API where you might get Token.
		type exampleOAuthClient struct {
			*AbstractClient
		}

		// AudiencePayload implementation it's what you only need to implement for oauth.AbstractClient.
		func (c *exampleOAuthClient) AudiencePayload() ([]byte, error) {
			reqBody, marshalErr := json.Marshal(map[string]string{
				"grant_type":    "client_credentials",         // depends on server settings
				"client_id":     "your_application_id",        // your client secret
				"client_secret": "your_application_secret",    // your client secret
				"audience":      "your_audience_of_interest",  // if needed
			})

			if marshalErr != nil {
				return nil, fmt.Errorf("error preparing audience payload: %w", marshalErr)
			}

			return reqBody, nil
		}
	*/

	// Create new instance.
	client := new(exampleOAuthClient)
	// Create new abstract client.
	client.AbstractClient = NewClient(&ClientConfig{AccessTokenURL: "https://login.mydomain.example/oauth/token"})
	client.AbstractClient.ClientOAuth = client

	// Then just call abstract method `AccessToken()` to get your access token and write needed code to handle it.
	client.AccessToken() // returns newToken, tokenErr
}
