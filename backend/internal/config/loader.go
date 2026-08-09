package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gconfig "github.com/raykavin/gobox/config"
)

// Options controls how the configuration is loaded.
type Options struct {
	// Path is the config file path (e.g. "/configs/api.yml").
	Path string

	// Defaults, when non-nil, is applied to the decoded config before
	// validation, filling in zero-valued optional fields.
	Defaults func(*Config)

	// Validate, when non-nil, checks cross-field invariants that
	// mapstructure cannot express. Each entry point supplies its own,
	// built from the ValidateX pieces below via Compose, since which
	// sections are actually required depends on which binary is loading
	// the file (e.g. cmd/api needs Server/OIDC, cmd/worker doesn't).
	// A nil Validate skips validation entirely.
	Validate func(*Config) error
}

// Load reads, applies defaults to, and validates the configuration file
// at opts.Path. It returns an error if the file is missing, malformed,
// or fails validation. The returned Config is safe to share and must not
// be mutated.
func Load(opts Options) (*Config, error) {
	if _, err := os.Stat(opts.Path); err != nil {
		return nil, fmt.Errorf("config file %q: %w", opts.Path, err)
	}

	ext := filepath.Ext(opts.Path)
	if ext == "" {
		return nil, fmt.Errorf("config file %q: missing extension", opts.Path)
	}

	lopts := gconfig.DefaultLoaderOptions[Config]()
	lopts.ConfigName = strings.TrimSuffix(filepath.Base(opts.Path), ext)
	lopts.ConfigType = ext[1:] // "yml"
	lopts.ConfigPaths = []string{filepath.Dir(opts.Path)}

	var cfg *Config
	var err error

	loader := gconfig.NewViper(lopts)
	if vFn := opts.Validate; vFn != nil {
		cfg, err = loader.LoadWithValidation(vFn)
	} else {
		cfg, err = loader.Load()
	}

	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if opts.Defaults != nil {
		opts.Defaults(cfg)
	}

	return cfg, nil
}

// Compose combines several ValidateX functions into one, running them in
// order and stopping at the first error. Build the Validate an entry point
// passes to Load from this, e.g.:
//
// config.Compose(config.ValidateApplication, config.ValidateServer,
// config.ValidateOIDC, config.ValidateDatabases)
func Compose(validators ...func(*Config) error) func(*Config) error {
	return func(c *Config) error {
		for _, v := range validators {
			if v == nil {
				continue
			}
			if err := v(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// ValidateApplication checks Application's required fields. Every entry
// point needs this.
func ValidateApplication(c *Config) error {
	if c.Application.Name == "" {
		return fmt.Errorf("application.name is required")
	}
	return nil
}

// ValidateServer checks Server's required fields. Only entry points that
// run an HTTP server need this (cmd/api, cmd/web).
func ValidateServer(c *Config) error {
	if c.Server.ListenPort == 0 {
		return fmt.Errorf("server.listen_port is required")
	}
	if tls := c.Server.TLS; tls != nil {
		if tls.Certificate == "" {
			return fmt.Errorf("server.tls.certificate is required when tls is set")
		}
		if tls.Key == "" {
			return fmt.Errorf("server.tls.key is required when tls is set")
		}
	}
	return nil
}

// ValidateOIDC checks that OIDC is configured. Only entry points that
// authenticate requests need this.
func ValidateOIDC(c *Config) error {
	if c.OIDC.IssuerURL == "" {
		return fmt.Errorf("oidc.issuer_url is required")
	}
	return nil
}

// ValidateOIDCFlow checks the additional OIDC fields required to run the
// server-side Authorization Code login flow (see
// internal/adapter/inbound/http/handler/auth.go). Only cmd/api needs this;
// ValidateOIDC alone is enough for a process that only verifies tokens.
func ValidateOIDCFlow(c *Config) error {
	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("oidc.client_secret is required")
	}
	if c.OIDC.RedirectURI == "" {
		return fmt.Errorf("oidc.redirect_uri is required")
	}
	if c.OIDC.PostLoginRedirectURI == "" {
		return fmt.Errorf("oidc.post_login_redirect_uri is required")
	}
	if c.OIDC.PostLogoutRedirectURI == "" {
		return fmt.Errorf("oidc.post_logout_redirect_uri is required")
	}
	return nil
}

// ValidateDatabases checks that the local database is configured. Every
// entry point that touches the database needs this.
func ValidateDatabases(c *Config) error {
	if c.Databases.Local.DSN == "" {
		return fmt.Errorf("databases.local.dsn is required")
	}
	return nil
}
