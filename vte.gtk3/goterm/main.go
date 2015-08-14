package main

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/sqp/vte/vte.gtk3"

	"os"
)

// Terminal settings.
const (
	TermFont  = "monospace 8"
	WinWidth  = 400
	WinHeight = 300
)

// Terminal command.
var TermCommand = []string{"bash", "-c", "echo ptour"}

func main() {
	gtk.Init(&os.Args)
	term, win, e := RunTerminal(TermCommand...)
	if e != nil {
		println(e.Error())
		return
	}

	// Settings.
	term.SetFontFromString(TermFont)
	win.SetSizeRequest(WinWidth, WinHeight)

	gtk.Main()
}

// RunTerminal runs the given command in a new terminal window.
//
func RunTerminal(args ...string) (*vte.Terminal, *gtk.Window, error) {
	terminal, window, e := vte.NewTerminalWindow()
	if e != nil {
		return nil, nil, e
	}

	// Signals.
	terminal.Connect("child-exited", gtk.MainQuit)
	window.Connect("destroy", gtk.MainQuit)

	// Start a command. This is optional, you can fill the terminal yourself.
	// See the package test for an example.
	if len(args) > 0 {
		e = terminal.Fork(args...)
		if e != nil {
			return nil, nil, e
		}
	}

	return terminal, window, nil
}
