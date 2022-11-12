package spinner

import (
	"time"

	"github.com/briandowns/spinner"
)

var characterSet = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var s *spinner.Spinner

func init() {
	s = spinner.New(characterSet, 100*time.Millisecond)
}

func Start(message string) {
	s.Suffix = " " + message

	if err := s.Color("cyan"); err != nil {
		// the only way we can arrive here is if cyan is not a valid color, which it is.
		panic(err)
	}

	s.Start()
}

func Stop() {
	s.Stop()
}
