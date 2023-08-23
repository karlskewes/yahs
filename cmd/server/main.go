package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/karlskewes/yahs"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("yahs starting")

	hs, err := yahs.New()
	if err != nil {
		log.Fatalf("failed to create new httpserver: %v", err)
	}

	app := NewApp(hs, logger)
	if err != nil {
		log.Fatalf("failed to create new app: %v", err)
	}

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
		return app.hs.Run(ctx)
	})

	if err := group.Wait(); err != nil {
		log.Print(err)
	}
}

type App struct {
	hs     *yahs.Server
	logger *slog.Logger
}

func NewApp(hs *yahs.Server, logger *slog.Logger) *App {
	app := &App{
		hs:     hs,
		logger: logger,
	}

	// attach routes to WebServer. This is a awkward compared to defining during
	// struct construction like `mc` but required in order for routes to have
	// access to private fields defined on the WebServer struct, such as loggers,
	// tracing, etc.
	app.addRoutes()

	return app
}

type GetConfigResponse struct{}

func (ws *App) addRoutes() {
	ws.hs.AddRoute("GET", "/", ws.Home)
}

func (ws *App) Home(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("home request received")
	_, err := w.Write([]byte(`yahs home page`))
	if err != nil {
		ws.logger.Error("failed to write body", err)
	}
}
