package main

import (
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

	err = env.TrackingWorker.Sync()
	if err != nil {
		fmt.Printf("ERROR: syncing tracked repositories: %v", err)
		os.Exit(1)
	}
}
