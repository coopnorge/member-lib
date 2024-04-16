package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
