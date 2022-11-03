package main

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var _ = goyek.Define(goyek.Task{
	Name:  "bump",
	Usage: "go get -u -t ./...",
	Action: func(tf *goyek.TF) {
		ForGoModules(tf, func(tf *goyek.TF) {
			cmd.Exec(tf, "go get -u -t ./...")
		}, dirBuild) // '/build' should be bumped without transitive dependencies as golangci-lint often has breaking changes
	},
})
