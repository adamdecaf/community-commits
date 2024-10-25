package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamdecaf/community-commits/internal/community_commits"
)

var (
	flagConfig = flag.String("config", os.Getenv("APP_CONFIG"), "Filepath to load config file from")
)

func main() {
	flag.Parse()

	env, err := community_commits.Setup(*flagConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = env.TrackingWorker.Sync()
	if err != nil {
		fmt.Printf("ERROR: syncing tracked repositories: %v", err)
		os.Exit(1)
	}

	// Setup app with webui and await shutdown
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Collect any errors
	errs := make(chan error)

	// Start our worker
	err = env.TrackingWorker.Start(ctx)
	if err != nil {
		errs <- fmt.Errorf("ERROR: starting tracked jobs: %v", err)
	}

	defer func() {
		err := env.TrackingWorker.Stop()
		if err != nil {
			env.Logger.Info("shutting down worker", slog.String("error", err.Error()))
		}
	}()

	// Listen for shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("signal: %v", <-c)
	}()

	err = <-errs
	if err != nil {
		env.Logger.Info(fmt.Sprintf("shutting down: %v", err))
	}
}
