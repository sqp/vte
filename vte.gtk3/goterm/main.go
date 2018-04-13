// Package goterm is an simple example creating a gtk/vte terminal window in go.
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

// TermCommand to run.
var TermCommand = []string{"top", "-n", "10"}

func main() {
	gtk.Init(&os.Args)

	term, win, e := vte.NewTerminalWindow()
	if e != nil {
		println(e.Error())
		return
	}

	// Settings.
	term.SetFontFromString(TermFont)
	win.SetSizeRequest(WinWidth, WinHeight)

	// Signals.
	term.Connect("child-exited", gtk.MainQuit)
	win.Connect("destroy", gtk.MainQuit)

	// Start a command. This is optional, you can fill the terminal yourself.
	// See the package test for an example.
	if len(TermCommand) > 0 {
		cmd := term.NewCmd(TermCommand...)
		cmd.OnExec = func(pid int, e error) {
			if e != nil {
				println(e)
				gtk.MainQuit()
			}
		}
		term.ExecAsync(cmd)
	}

	gtk.Main()
}
