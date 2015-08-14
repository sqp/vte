package vte_test

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	vtecommon "github.com/sqp/vte"
	"github.com/sqp/vte/vte.gtk3"

	"fmt"
	"os"
	"os/exec"
)

// Terminal settings.
const (
	TermFont            = "monospace 8"
	WinWidth, WinHeight = 400, 300

	bashCmd = "echo closing in 3 sec; sleep 1; echo 2; sleep 1; echo 1; sleep 1"
)

func Example_fork() {
	gtk.Init(&os.Args)

	term, win, e := vte.NewTerminalWindow()
	if e != nil {
		fmt.Println(e)
		return
	}

	// Settings.
	term.SetColorsFromStrings(vtecommon.MikePal)
	term.SetFontFromString(TermFont)
	win.SetSizeRequest(WinWidth, WinHeight)

	// Signals.
	win.Connect("destroy", gtk.MainQuit)

	e = term.Fork("sh", "-c", bashCmd)
	if e != nil {
		fmt.Println(e)
		return
	}

	// Wait our command completion as a GTK callback.
	term.Connect("child-exited", func() {
		// Get the terminal text using the clipboard and print the result for the test.
		str := getTerminalText(term)
		fmt.Print(str)

		win.Destroy() // this release the main loop and finish the test.
	})

	// Start the main loop to run the test.
	gtk.Main()

	// Output:
	// closing in 3 sec
	// 2
	// 1
}

func Example_execCmd() {
	gtk.Init(&os.Args)

	term, win, e := vte.NewTerminalWindow()
	if e != nil {
		fmt.Println(e)
		return
	}

	// Settings.
	term.SetColorsFromStrings(vtecommon.MikePal)
	term.SetFontFromString(TermFont)
	win.SetSizeRequest(WinWidth, WinHeight)

	// Signals.
	win.Connect("destroy", gtk.MainQuit)

	glib.IdleAdd(func() { // Wait gtk to be ready to start our command in the gtk loop.

		cmd := exec.Command("sh", "-c", bashCmd)
		cmd.Stdout = term // use the terminal as the command output (writer)
		cmd.Stderr = term

		e = cmd.Start()
		if e != nil {
			fmt.Println(e)
			return
		}

		// Use a go routine to release the gtk main loop and wait our command completion.
		go func() {
			cmd.Wait()

			glib.IdleAdd(func() { // Resynced in the GTK loop to prevent thread crashes.

				// Get the terminal text using the clipboard and print the result for the test.
				str := getTerminalText(term)
				fmt.Print(str)

				win.Destroy() // this release the main loop and finish the test.
			})
		}()
	})

	// Start the main loop to run the test.
	gtk.Main()

	// Output:
	// closing in 3 sec
	// 2
	// 1
}

func getTerminalText(term *vte.Terminal) string {
	term.SelectAll()
	term.CopyClipboard()

	clip, e := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if e != nil {
		return e.Error()
	}
	str, e := clip.WaitForText()
	if e != nil {
		return e.Error()
	}
	return str
}
