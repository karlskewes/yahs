package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	httpserver "github.com/karlskewes/go-yahs"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.Print("go-yahs starting")

	app := httpserver.NewApp()

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
