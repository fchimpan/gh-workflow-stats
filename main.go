package main

import (
	"github.com/fchimpan/gh-workflow-stats/workflow"
)

func main() {
	if err := workflow.Execute(); err != nil {
		panic(err)
	}
}
