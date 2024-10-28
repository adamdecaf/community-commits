package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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

	// Setup app and await shutdown
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Start our worker
	err = env.TrackingWorker.Start(ctx)
	if err != nil {
		env.Logger.Info(fmt.Sprintf("shutting down: %v", err))
		os.Exit(1)
	}
}
