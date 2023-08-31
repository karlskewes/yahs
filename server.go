package yahs

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	httpShutdownPreStopDelaySeconds = 1
	httpShutdownTimeoutSeconds      = 1
	defaultListenAddr               = "localhost:8080"
)

// Route contains the HTTP routing information required to match an incoming
// HTTP request and direct to the included handler.
type Route struct {
	method  string
	pattern *regexp.Regexp
	handler http.HandlerFunc
}

// NewRoute creates a HTTP Route which can be used with AddRoute() or SetRoute().
// Regex pattern input will be anchored.
func NewRoute(method, pattern string, handler http.HandlerFunc) Route {
	return Route{
		method:  method,
		pattern: regexp.MustCompile("^" + pattern + "$"),
		handler: handler,
	}
}

// Server contains configuration for the HTTP Server.
type Server struct {
	routes    []Route
	srv       *http.Server
	assets    http.FileSystem
	templates map[string]*template.Template
}

// Option configures a HTTP Server.
type Option func(hs *Server) error

// New returns an instance of our HTTP Server with default settings.
// Custom configuration provided by supplying Options take precedence.
func New(options ...Option) (*Server, error) {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    defaultListenAddr,
		Handler: mux,
	}

	hs := &Server{
		srv:       srv,
		templates: make(map[string]*template.Template),
	}

	// Delegate HTTP routing to our custom ServeHTTP handler Serve().
	// This is a awkward compared to defining during struct construction like
	// `srv` but required in order for Serve to have access to private fields
	// on the Server struct like `hs.routes`.
	mux.Handle("/", http.HandlerFunc(hs.Serve))

	for _, option := range options {
		err := option(hs)
		if err != nil {
			return nil, err
		}
	}

	return hs, nil
}

// WithListenAddress Option enables supplying a custom listen address in the
// format "<ip>:<port>".
func WithListenAddress(listenAddr string) Option {
	return func(hs *Server) error {
		if listenAddr == "" {
			return errors.New("provided listen address must not be empty")
		}

		hs.srv.Addr = listenAddr

		return nil
	}
}

// WithHTTPServer Option enables supplying a custom http.Server configured with
// handler, timeouts, listen address, transport configuration, etc.
func WithHTTPServer(srv *http.Server) Option {
	return func(hs *Server) error {
		if srv == nil {
			return errors.New("provided http.Server must not be nil")
		}

		hs.srv = srv

		return nil
	}
}

// AddRoute adds a route to the HTTP mux. Routes are evaluated sequentially
// until the first match is found.
func (hs *Server) AddRoute(route Route) {
	// TODO: race condition, mutating slice whilst Serve could be reading slice.
	// Consider locking during hs.Run().
	hs.routes = append(hs.routes, route)
}

// SetRoutes replaces any existing routes with the slice of routes provided.
// Routes are evaluated sequentially until the first match is found.
func (hs *Server) SetRoutes(routes []Route) {
	if routes == nil {
		routes = make([]Route, 0)
	}

	hs.routes = routes
}

// Serve is the entry point for handling all http requests. It performs routing
// of the request to the appropriate handler if any.
// Exported method to enable easier testing in consumer packages.
func (hs *Server) Serve(w http.ResponseWriter, r *http.Request) {
	// keep track of allowed methods for the Request path (if any)
	var allow []string

	// check if the Request path matches any handled routes
	for _, route := range hs.routes {
		matches := route.pattern.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				// the current path match has a different method, append to list for
				// returning if no matching method is found
				allow = append(allow, route.method)
				continue
			}

			// found matching path & method, pass to handler
			route.handler(w, r)
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r) // Could replace with a specific not found handler
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
		drainTimeoutCtx, cancel := context.WithTimeout(ctx, httpShutdownTimeoutSeconds*time.Second)
		defer cancel()

		return hs.srv.Shutdown(drainTimeoutCtx)
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
