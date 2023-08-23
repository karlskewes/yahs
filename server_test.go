package yahs

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const httpListenPort = 65123

func TestNewHTTPServer(t *testing.T) {
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

func TestRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpListenPort),
		Handler: mux,
	}

	hs, err := New(WithHTTPServer(srv))
	if err != nil {
		t.Errorf("failed to set legitimate http server: %v", err)
	}

	//nolint:errcheck // If app fails to run then the test will fail.
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
