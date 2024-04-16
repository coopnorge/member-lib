<!-- markdownlint-disable-file MD009 -->
# OAuth

This package contains abstract client that allows work with OAuth 2 to obtain
Access token.

If you use this package in your project you need implement concrete
implementation of it.

Simple example:

```go
package example

import (
  "encoding/json"
  "fmt"

  "github.com/coopnorge/member-lib/security/oauth"
)

type OAuthClient struct {
  *oauth.AbstractClient
}

// NewOAuthClient constructor.
func NewOAuthClient() *OAuthClient {
  client := new(OAuthClient)
  client.AbstractClient = oauth.NewClient(&oauth.ClientConfig{AccessTokenURL: "https://login.mydomain.example/oauth/token"})
  client.AbstractClient.ClientOAuth = client

  return client
}

// AudiencePayload implementation it's what you only need to implement for oauth.AbstractClient.
func (c *OAuthClient) AudiencePayload() ([]byte, error) {
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

```

Then just call abstract method `AccessToken()` to get your access token and
write needed code to handle it. 
