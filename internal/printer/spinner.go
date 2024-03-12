package printer

import (
	"time"

	"github.com/briandowns/spinner"
)

type Spinner spinner.Spinner

type SpinnerOptions struct {
	Text          string
	CharSetsIndex int
	Color         string
}

func NewSpinner(opt SpinnerOptions) *Spinner {
	s := spinner.New(spinner.CharSets[opt.CharSetsIndex], 100*time.Millisecond)
	s.Suffix = opt.Text
	_ = s.Color(opt.Color)
	return (*Spinner)(s)
}

func (s *Spinner) Start() {
	(*spinner.Spinner)(s).Start()
}

func (s *Spinner) Stop() {
	(*spinner.Spinner)(s).Stop()
}

func (s *Spinner) Update(opt SpinnerOptions) {
	(*spinner.Spinner)(s).Suffix = opt.Text
}
