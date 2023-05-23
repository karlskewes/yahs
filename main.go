package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	httpShutdownPreStopDelaySeconds = 1
	httpShutdownTimeoutSeconds      = 1
)

func main() {
	log.Print("go-yahs starting")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var group errgroup.Group

	group.Go(func() error {
		<-ctx.Done()

		log.Print("received OS signal to shutdown, use Ctrl+C again to force")

		// reset signals so a second ctrl+c will terminate the application.
		stop()

		return nil
	})

	group.Go(func() error {
		return Run(ctx)
	})

	if err := group.Wait(); err != nil {
		log.Print(err)
	}
}

// Run starts an HTTP server and gracefully shuts down when the provided
// context is marked done.
func Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())
	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

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

		return srv.Shutdown(ctx2)
	})

	group.Go(func() error {
		err := srv.ListenAndServe()
		// http.ErrServerClosed is expected at shutdown.
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	})

	return group.Wait()
}
