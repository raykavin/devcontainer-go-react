package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"my-app/internal/port"
)

// server abstracts what Run needs: something that listens, reports its
// address, and drains gracefully. Both *http.Server and *gobox.Engine
// satisfy this but note *http.Server does NOT serve TLS via ListenAndServe,
// so wrap it (see httpServer) if you need SSL dispatch.
type server interface {
	Listen() error
	Addr() string
	Shutdown(ctx context.Context) error
}

// Run serves srv until ctx is cancelled, then drains within shutdownTO.
// Returns only after shutdown completes.
func Run(ctx context.Context, srv server, shutdownTO time.Duration, log port.Logger, serverName ...string) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Listen() }()

	l := log.WithField("addr", srv.Addr())
	if len(serverName) > 0 {
		l = l.WithField("server_name", serverName[0])
	}
	l.Info("http server started")

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTO)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

// HTTPServer adapts *http.Server to the server interface. It uses
// ListenAndServeTLS when a TLSConfig is present, since ListenAndServe
// never serves TLS on its own.
type HTTPServer struct{ *http.Server }

func NewHTTPServer(server *http.Server) server {
	return HTTPServer{Server: server}
}

var _ server = (*HTTPServer)(nil)

func (s HTTPServer) Listen() error {
	if s.TLSConfig != nil && (len(s.TLSConfig.Certificates) > 0 || s.TLSConfig.GetCertificate != nil) {
		return s.ListenAndServeTLS("", "")
	}
	return s.ListenAndServe()
}

func (s HTTPServer) Addr() string { return s.Server.Addr }
