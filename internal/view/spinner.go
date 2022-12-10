package view

var characterSet = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (view *View) StartSpinner(message string) {
	if !view.isInteractive {
		return
	}

	view.spinner.Suffix = " " + message

	if err := view.spinner.Color("cyan"); err != nil {
		// the only way we can arrive here is if cyan is not a valid color, which it is.
		panic(err)
	}

	view.spinner.Start()
}

func (view *View) StopSpinner() {
	if !view.isInteractive {
		return
	}
	view.spinner.Stop()
}
