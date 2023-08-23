package yahs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	httpShutdownPreStopDelaySeconds = 1
	httpShutdownTimeoutSeconds      = 1
	defaultListenAddr               = "localhost:8080"
)

// Server holds configuration for our httpserver.
type Server struct {
	srv *http.Server
}

// Option configures a HTTP Server.
type Option func(hs *Server) error

// New returns an instance of our httpserver with default settings.
// Custom configuration provided by supplying Options take precedence.
func New(options ...Option) (*Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())
	srv := &http.Server{
		Addr:    defaultListenAddr,
		Handler: mux,
	}

	hs := &Server{
		srv: srv,
	}

	for _, option := range options {
		err := option(hs)
		if err != nil {
			return nil, err
		}
	}

	return hs, nil
}

// WithHTTPServer Option enables supplying a custom http.Server configured with
// handler, timeouts, listen address, transport configuration, etc.
func WithHTTPServer(srv *http.Server) Option {
	return func(hs *Server) error {
		if srv == nil {
			return fmt.Errorf("provided http.Server must not be nil")
		}

		hs.srv = srv

		return nil
	}
}

// Run starts the HTTP Server application and gracefully shuts down when the
// provided context is marked done.
func (hs *Server) Run(ctx context.Context) error {
	var group errgroup.Group

	group.Go(func() error {
		<-ctx.Done()

		// before shutting down the HTTP server wait for any HTTP requests that are
		// in transit on the network. Common in Kubernetes and other distributed
		// systems.
		time.Sleep(httpShutdownPreStopDelaySeconds * time.Second)

		// give active connections time to complete or disconnect before closing.
		ctx2, cancel := context.WithTimeout(ctx, httpShutdownTimeoutSeconds*time.Second)
		defer cancel()

		return hs.srv.Shutdown(ctx2)
	})

	group.Go(func() error {
		err := hs.srv.ListenAndServe()
		// http.ErrServerClosed is expected at shutdown.
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	})

	return group.Wait()
}
