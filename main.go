package main

import (
	"github.com/fchimpan/gh-workflow-stats/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
