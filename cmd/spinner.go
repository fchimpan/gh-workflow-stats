package cmd

import (
	"time"

	"github.com/briandowns/spinner"
)

const (
	spinnerText = "  fetching workflow runs..."
)

type spinnerOptions struct {
	text          string
	charSetsIndex int
	color         string
}

func newSpinner(opt spinnerOptions) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[opt.charSetsIndex], 100*time.Millisecond)
	s.Suffix = opt.text
	s.Color(opt.color)
	return s
}
