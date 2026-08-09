package config

import (
	"time"
)

// Application holds application-level settings such as identity,
// versioning, and observability.
type Application struct {
	// Name is the application name used for logging, telemetry,
	// and service identification.
	Name string `mapstructure:"name"`

	// Description is a short human-readable description of the application.
	Description string `mapstructure:"description"`

	// Version is the application version (e.g. semantic version or
	// build identifier).
	Version string `mapstructure:"version"`

	// LogLevel is the application log level
	// (e.g. "debug", "info", "warn", "error").
	LogLevel string `mapstructure:"log_level"`

	// HealthServerListenPort is the TCP port on which the health check
	// server listens.
	HealthServerListenPort uint16 `mapstructure:"health_server_listen_port"`
}

// TLS holds TLS certificate settings. Its presence on HTTPServer enables
// TLS; a validated non-nil TLS always carries usable paths.
type TLS struct {
	// Certificate is the path to the TLS certificate file.
	Certificate string `mapstructure:"certificate"`

	// Key is the path to the TLS private key file.
	Key string `mapstructure:"key"`
}

// HTTPServer holds HTTP server settings, including TLS, routing fallback,
// and connection timeouts.
type HTTPServer struct {
	// ListenPort is the TCP port on which the HTTP server listens.
	ListenPort uint16 `mapstructure:"listen_port"`

	// TLS enables HTTPS when non-nil. A nil value means plain HTTP.
	// After Load, a non-nil TLS is guaranteed to have both Certificate
	// and Key set.
	TLS *TLS `mapstructure:"tls"`

	// NoRouteRedir is the URL to which unmatched routes are redirected.
	// An empty string disables redirection.
	NoRouteRedir string `mapstructure:"no_route_redir"`

	// WriteTimeout is the maximum duration before timing out writes
	// of the response.
	WriteTimeout time.Duration `mapstructure:"write_timeout"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout time.Duration `mapstructure:"read_timeout"`

	// IdleTimeout is the maximum amount of time to wait for the next
	// request when keep-alives are enabled.
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

// TLSEnabled reports whether the server is configured to serve over TLS.
func (s HTTPServer) TLSEnabled() bool { return s.TLS != nil }

// OIDC holds OpenID Connect client settings used for token validation,
// authorization, and the server-side Authorization Code login flow (see
// internal/infra/oidc/flow.go and internal/adapter/inbound/http/handler/auth.go).
// The backend acts as a confidential client: it performs the browser
// redirect, exchanges the code for tokens itself, and hands the browser
// only HttpOnly cookies the access/refresh/ID tokens never reach
// frontend JavaScript.
type OIDC struct {
	// IssuerURL is the base URL of the OIDC issuer
	// (e.g. a Keycloak realm URL).
	IssuerURL string `mapstructure:"issuer_url"`

	// ClientID is the OAuth2 client identifier registered with the issuer.
	ClientID string `mapstructure:"client_id"`

	// ClientSecret is the OAuth2 client secret. Required for confidential
	// clients, token introspection, and the login/refresh code exchange.
	ClientSecret string `mapstructure:"client_secret"`

	// DisableIntrospection indicates whether remote token introspection
	// should be skipped, falling back to local JWT validation only.
	DisableIntrospection bool `mapstructure:"disable_introspection"`

	// RedirectURI is this backend's own OAuth2 callback URL, registered as
	// a "Valid Redirect URI" on the issuer's client (e.g.
	// https://api.example.com/api/v1/auth/callback). The browser is sent
	// here directly by the issuer; it is never the frontend's URL.
	RedirectURI string `mapstructure:"redirect_uri"`

	// Scopes are the OIDC scopes requested during login. Defaults to
	// {"openid", "profile", "email"} when empty.
	Scopes []string `mapstructure:"scopes"`

	// PostLoginRedirectURI is the frontend URL the browser lands on after
	// a successful login callback when no (or an invalid) redirect_uri was
	// supplied to /api/v1/auth/login.
	PostLoginRedirectURI string `mapstructure:"post_login_redirect_uri"`

	// PostLogoutRedirectURI is the frontend URL the issuer sends the
	// browser back to after ending the SSO session on logout.
	PostLogoutRedirectURI string `mapstructure:"post_logout_redirect_uri"`

	// CookieDomain is the Domain attribute set on auth cookies. Empty
	// leaves it unset (host-only cookie).
	CookieDomain string `mapstructure:"cookie_domain"`

	// CookiePath is the Path attribute set on every auth cookie (access,
	// refresh, ID, CSRF, and the short-lived login-flow cookies). Defaults
	// to "/" when empty. A single shared path keeps cookie scoping correct
	// regardless of any reverse-proxy path prefix in front of this API
	// (see cmd/api/main.go's @BasePath) narrower per-purpose paths would
	// need to account for that prefix, which this backend cannot see.
	CookiePath string `mapstructure:"cookie_path"`

	// CookieSecure sets the Secure attribute on auth cookies. Should be
	// true everywhere except plain-HTTP local development.
	CookieSecure bool `mapstructure:"cookie_secure"`

	// CookieSameSite sets the SameSite attribute on auth cookies: "lax"
	// (default), "strict", or "none". Lax is required for the state/verifier
	// cookies to survive the issuer's cross-site redirect back to
	// /api/v1/auth/callback.
	CookieSameSite string `mapstructure:"cookie_same_site"`
}

// DatabaseSettings holds a single database connection's settings.
type DatabaseSettings struct {
	// DSN is the database connection string.
	DSN string `mapstructure:"dsn"`

	// LogLevel is the database driver's log level.
	LogLevel string `mapstructure:"log_level"`

	// Dialector is the GORM dialector name (e.g. "postgres", "mysql").
	Dialector string `mapstructure:"dialector"`
}

// Databases groups the application's database connections.
type Databases struct {
	// Local is the configuration for the application's own database.
	Local DatabaseSettings `mapstructure:"local"`
}

// CORS holds the settings for the CORS middleware (see
// internal/adapter/inbound/http/middleware/cors.go). Required once cookie
// based auth is in use since Access-Control-Allow-Origin: * is incompatible
// with credentialed requests browsers reject it outright, so origins must
// be enumerated explicitly.
type CORS struct {
	// AllowedOrigins lists the exact origins (scheme + host + port) allowed
	// to make credentialed requests to this API. An origin not in this list
	// receives no CORS headers, so the browser blocks the response.
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// Config is the root application configuration, aggregating all
// configuration sections. It is populated by Load and must not be
// mutated afterwards.
type Config struct {
	// Application contains application-level settings.
	Application Application `mapstructure:"application"`

	// Server contains HTTP server settings.
	Server HTTPServer `mapstructure:"server"`

	// OIDC contains OpenID Connect settings.
	OIDC OIDC `mapstructure:"oidc"`

	// Databases contains the database connections settings.
	Databases Databases `mapstructure:"databases"`

	// CORS contains the cross-origin settings for the HTTP API.
	CORS CORS `mapstructure:"cors"`
}
