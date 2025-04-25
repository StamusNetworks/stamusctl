package logging

import (
	"time"

	"github.com/briandowns/spinner"
)

func NewSpinner(prefix, finalMSG string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond)
	s.Prefix = prefix
	s.FinalMSG = finalMSG

	s.Start()

	return s
}

func SpinnerStop(s *spinner.Spinner) {
	if s != nil {
		s.Stop()
	}
}
