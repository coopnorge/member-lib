// Package oauth provides simple abstract client that allows work with OAuth 2 to obtain Access token.
package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Token represents the structure of the OAuth 2.0 token response.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
}

// ClientConfig that will be applied to Client.
type ClientConfig struct {
	// AccessTokenURL authorization server.
	AccessTokenURL string
	// Transport that supports HTTP protocol.
	Transport *http.Client
}

// ClientOAuth interface of Client.
type ClientOAuth interface {
	// AudiencePayload must return json payload that will be sent in request body to get Token when called AccessToken.
	AudiencePayload() ([]byte, error)
	// AccessToken allows to obtain access Token that is registered for the application client.
	AccessToken() (Token, error)
}

// AbstractClient allows get access token from IDP services.
type AbstractClient struct {
	acTokenURL  string
	transport   *http.Client
	cachedToken Token
	mu          sync.RWMutex
	ClientOAuth
}

// NewClient that allows get access token.
func NewClient(cfg *ClientConfig) *AbstractClient {
	c := &AbstractClient{acTokenURL: cfg.AccessTokenURL}

	if cfg.Transport != nil {
		c.transport = cfg.Transport
	} else {
		c.transport = &http.Client{Timeout: time.Minute}
	}

	return c
}

// AudiencePayload must return json payload that will be sent in request body to get Token when called AccessToken.
func (c *AbstractClient) AudiencePayload() ([]byte, error) {
	panic("must be not implement in AbstractClient")
}

// AccessToken allows to obtain access token that is registered for the application client.
func (c *AbstractClient) AccessToken() (Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isValidToken() {
		return c.cachedToken, nil
	}
	newToken, err := c.getNewAccessToken()
	if err != nil {
		return Token{}, err
	}
	c.cachedToken = newToken
	return newToken, nil
}

func (c *AbstractClient) isValidToken() bool {
	if c.cachedToken.AccessToken == "" {
		return false
	}
	expiry := time.Now().Add(time.Duration(c.cachedToken.ExpiresIn) * time.Second)
	return !time.Now().After(expiry)
}

func (c *AbstractClient) getNewAccessToken() (Token, error) {
	payload, payloadErr := c.ClientOAuth.AudiencePayload()
	if payloadErr != nil {
		return Token{}, payloadErr
	}

	req, newReqErr := http.NewRequestWithContext(context.Background(), http.MethodPost, c.acTokenURL, bytes.NewBuffer(payload))
	if newReqErr != nil {
		return Token{}, fmt.Errorf("error creating request: %w", newReqErr)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, sendReqErr := c.transport.Do(req)
	if sendReqErr != nil {
		return Token{}, fmt.Errorf("error sending request: %w", sendReqErr)
	}
	defer func() {
		respBodyCloseErr := resp.Body.Close()
		if respBodyCloseErr != nil {
			log.Printf("error closing response body: %v", respBodyCloseErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, bodyReadErr := io.ReadAll(resp.Body)
		if bodyReadErr != nil {
			return Token{}, fmt.Errorf("error parsing body of the response: %w", bodyReadErr)
		}
		return Token{}, fmt.Errorf("token request failed: %s", body)
	}

	var tokenResponse Token
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return Token{}, fmt.Errorf("error decoding response: %w", err)
	}
	return tokenResponse, nil
}
