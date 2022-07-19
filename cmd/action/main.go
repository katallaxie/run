package main

import (
	a "github.com/katallaxie/run/pkg/action"

	githubactions "github.com/sethvargo/go-githubactions"
)

var action *githubactions.Action

func run() error {
	action = githubactions.New()

	_, err := a.NewFromInputs(action)
	if err != nil {
		return nil
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		action.Fatalf("%v", err)
	}
}
