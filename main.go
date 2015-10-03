package main

import (
	"fmt"

	"github.com/google/go-github/github"
)

func rebased(client *github.Client, owner, repo, base, head string) (bool, error) {
	comp, _, err := client.Repositories.CompareCommits(owner, repo, base, head)
	if err != nil {
		return false, fmt.Errorf("Error during comparing commits: %s", err)
	}

	return *comp.BehindBy == 0, nil
}

func main() {
	client := github.NewClient(nil)

	owner := "grpc"
	repo := "grpc-go"
	base := "grpc:master"
	head := "mwitkow-io:monitoring_take_i"

	ok, err := rebased(client, owner, repo, base, head)
	if err != nil {
		panic(err)
	}

	if ok {
		fmt.Println("ALL GOOD")
	} else {
		fmt.Println("need to rebase!")
	}
}
