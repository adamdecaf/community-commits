package main

import (
	"fmt"

	"github.com/adamdecaf/community-commits/internal/community_commits"
)

func main() {
	env := community_commits.Setup()

	fmt.Printf("env: %#v\n", env)
}
