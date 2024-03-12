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

func NewSpinner(opt SpinnerOptions) (*Spinner, error) {
	s := spinner.New(spinner.CharSets[opt.CharSetsIndex], 100*time.Millisecond)
	s.Suffix = opt.Text
	if err := s.Color(opt.Color); err != nil {
		return nil, err
	}
	return (*Spinner)(s), nil
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
