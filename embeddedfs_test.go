package yahs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithEmbeddedFS(t *testing.T) {
	t.Parallel()
	wwwroot := NewWWWRoot()
	hs, err := New(WithEmbeddedFS(wwwroot))
	if err != nil {
		t.Fatalf("failed to create http server: %v", err)
	}

	if hs.assets == nil {
		t.Error("asset filesystem was not created")
	}
	if hs.templates == nil {
		t.Error("templates map was not created")
	}
}

func TestHandleStaticFiles(t *testing.T) {
	t.Parallel()

	wwwroot := NewWWWRoot()
	hs, err := New(WithEmbeddedFS(wwwroot))
	if err != nil {
		t.Fatalf("failed to create http server: %v", err)
	}

	testCases := []struct {
		description string
		path        string
		statusCode  int
	}{
		{"file exists", "/static/favicon.ico", http.StatusOK},
		{"invalid path", "/static/path/to/non-existant/static-file", http.StatusNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			hs.Serve(w, req)
			res := w.Result()
			defer res.Body.Close()
			if tc.statusCode != res.StatusCode {
				t.Errorf("want: %d, got: %d", tc.statusCode, res.StatusCode)
			}
		})
	}
}

func TestHandleTemplates(t *testing.T) {
	t.Parallel()

	wwwroot := NewWWWRoot()
	hs, err := New(WithEmbeddedFS(wwwroot))
	if err != nil {
		t.Fatalf("failed to create http server: %v", err)
	}

	testCases := []struct {
		description string
		path        string
		statusCode  int
	}{
		{"valid path", "/index.html", http.StatusOK},
		{"invalid path", "/path/to/non-existant/templated-file", http.StatusNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			hs.handleTemplates().ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			if tc.statusCode != res.StatusCode {
				t.Errorf("want: %d, got: %d", tc.statusCode, res.StatusCode)
			}
		})
	}
}

func BenchmarkHandleTemplates(b *testing.B) {
	wwwroot := NewWWWRoot()
	hs, err := New(WithEmbeddedFS(wwwroot))
	if err != nil {
		b.Fatalf("failed to create http server: %v", err)
	}

	// TODO
	// b.Cleanup(func() { _ = s.Shutdown() })
	b.Logf("b.N is %d", b.N)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		w := httptest.NewRecorder()
		hs.handleTemplates().ServeHTTP(w, req)
	}
}
