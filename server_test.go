package yahs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const httpListenPort = 65123

func TestNew(t *testing.T) {
	t.Parallel()
	_, err := New()
	if err != nil {
		t.Errorf("failed to create new http server: %v", err)
	}
}

func TestWithHTTPServer(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    defaultListenAddr,
		Handler: mux,
	}

	hs, err := New(WithHTTPServer(srv))
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	if hs.srv != srv {
		t.Errorf("unexpected http.Server set, want: %v - got: %v", srv, hs.srv)
	}
}

func TestWithListenAddr(t *testing.T) {
	t.Parallel()
	want := "1.2.3.4:8080"

	hs, err := New(WithListenAddress(want))
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	if hs.srv.Addr != want {
		t.Errorf("unexpected http.Addr set, want: %v - got: %v", want, hs.srv.Addr)
	}
}

func TestAddRoute(t *testing.T) {
	t.Parallel()

	hs, err := New()
	if err != nil {
		t.Errorf("failed to create new http server: %v", err)
	}

	method := "GET"
	pattern := "/path/to/endpoint"
	handler := http.NotFoundHandler().ServeHTTP

	route := NewRoute(method, pattern, handler)
	hs.AddRoute(route)

	if len(hs.routes) != 1 {
		t.Errorf("unexpected number of routes, want: %d - got: %d", 1, len(hs.routes))
	}

	if hs.routes[0].method != method {
		t.Errorf("unexpected method, want: %s - got: %s", method, hs.routes[0].method)
	}

	if hs.routes[0].pattern.String() != "^"+pattern+"$" {
		t.Errorf("unexpected regex pattern, want: %s - got: %s", pattern, hs.routes[0].pattern.String())
	}
}

func TestSetRoute(t *testing.T) {
	t.Parallel()

	hs, err := New()
	if err != nil {
		t.Errorf("failed to create new http server: %v", err)
	}

	method := "GET"
	pattern := "/path/to/endpoint"
	handler := http.NotFoundHandler().ServeHTTP

	route1 := NewRoute(method, pattern+"1", handler)
	route2 := NewRoute(method, pattern+"2", handler)
	hs.SetRoutes([]Route{route1, route2})

	if len(hs.routes) != 2 {
		t.Errorf("unexpected number of routes, want: %d - got: %d", 2, len(hs.routes))
	}

	if hs.routes[0].pattern.String() != "^"+pattern+"1$" {
		t.Errorf("unexpected regex pattern, want: %s - got: %s", pattern+"1", hs.routes[0].pattern.String())
	}

	if hs.routes[1].pattern.String() != "^"+pattern+"2$" {
		t.Errorf("unexpected regex pattern, want: %s - got: %s", pattern+"2", hs.routes[1].pattern.String())
	}
}

func TestServe_NoRoutes(t *testing.T) {
	hs, err := New()
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	ts := httptest.NewServer(hs.srv.Handler)

	t.Cleanup(func() {
		ts.Close()
	})

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("failed to query http server: %v", err)
	}

	want := http.StatusNotFound
	if resp.StatusCode != want {
		t.Errorf("want: %d - got: %d", want, resp.StatusCode)
	}
}

func TestServe_CustomRoute(t *testing.T) {
	hs, err := New()
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	method := "GET"
	pattern := "/path/to/endpoint"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}

	hs.AddRoute(
		NewRoute(method, pattern, handler),
	)

	ts := httptest.NewServer(hs.srv.Handler)

	t.Cleanup(func() {
		ts.Close()
	})

	resp, err := http.Get(ts.URL + pattern)
	if err != nil {
		t.Errorf("failed to query http server: %v", err)
	}

	want := http.StatusTeapot
	if resp.StatusCode != want {
		t.Errorf("want: %d - got: %d", want, resp.StatusCode)
	}
}

func TestRun(t *testing.T) {
	hs, err := New(
		WithListenAddress(fmt.Sprintf(":%d", httpListenPort)),
	)
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	//nolint:errcheck // If http server fails to run then the test will fail.
	go hs.Run(context.Background())

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("http://localhost:%d/", httpListenPort),
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create a new request: %v", err)
	}

	// HTTP server `go main()` goroutine might not be scheduled yet.
	// Attempt GET request a few times with a delay between each request.
	client := &http.Client{}
	var resp *http.Response
	var doErr error

	for i := 0; i < 3; i++ {
		resp, doErr = client.Do(req)
		if doErr == nil {
			defer resp.Body.Close()

			break
		}

		// wait for server to startup
		time.Sleep(time.Duration(i) * time.Second)
	}

	if doErr != nil {
		t.Fatalf("failed to query HTTP server")
	}

	want := http.StatusNotFound
	if resp.StatusCode != want {
		t.Errorf("want: %d - got: %d", want, resp.StatusCode)
	}
}
