package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/karlskewes/yahs"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.Print("yahs starting")

	app, err := yahs.NewApp()
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
		return app.Run(ctx)
	})

	if err := group.Wait(); err != nil {
		log.Print(err)
	}
}
