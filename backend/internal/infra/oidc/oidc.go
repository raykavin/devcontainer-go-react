package oidc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raykavin/gobox/oidcauth"
)

const (
	oidcCacheTTL       = 5 * time.Minute
	oidcRequestTimeout = 30 * time.Second
)

func NewMemoryCache(
	ctx context.Context,
	ttl time.Duration,
) oidcauth.Cache {
	if ttl == 0 {
		ttl = oidcCacheTTL
	}
	return oidcauth.NewMemoryCache(ctx, ttl)
}

// New creates an OIDC client with an in-memory token cache.
func New(
	ctx context.Context,
	config oidcauth.Config,
	opts ...oidcauth.Option,
) (*oidcauth.OIDC, error) {
	client, err := oidcauth.New(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("initialize oidc client: %w", err)
	}

	return client, nil
}

// NewFlow builds the server-side Authorization Code + PKCE flow.
func NewFlow(
	ctx context.Context,
	issuerURL string,
	clientID string,
	clientSecret string,
	redirectURI string,
	scopes []string,
) (*oidcauth.Flow, error) {
	if issuerURL == "" {
		return nil, errors.New("oidc flow: issuer url cannot be empty")
	}
	if clientID == "" {
		return nil, errors.New("oidc flow: client id cannot be empty")
	}
	if clientSecret == "" {
		return nil, errors.New("oidc flow: client secret cannot be empty")
	}
	if redirectURI == "" {
		return nil, errors.New("oidc flow: redirect uri cannot be empty")
	}
	if len(scopes) == 0 {
		return nil, errors.New("oidc flow: scopes cannot be empty")
	}

	return oidcauth.NewFlow(ctx, oidcauth.FlowConfig{
		IssuerURL:    issuerURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
	})
}
