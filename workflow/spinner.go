package workflow

import (
	"time"

	"github.com/briandowns/spinner"
)

type workflowSpinner spinner.Spinner

type spinnerOptions struct {
	text          string
	charSetsIndex int
	color         string
}

func newSpinner(opt spinnerOptions) *workflowSpinner {
	s := spinner.New(spinner.CharSets[opt.charSetsIndex], 100*time.Millisecond)
	s.Suffix = opt.text
	s.Color(opt.color)
	return (*workflowSpinner)(s)
}

func (s *workflowSpinner) start() {
	((*spinner.Spinner)(s)).Start()
}

func (s *workflowSpinner) stop() {
	((*spinner.Spinner)(s)).Stop()
}

func (s *workflowSpinner) update(opt spinnerOptions) {
	s.Suffix = opt.text
}
