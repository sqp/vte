package vte_test

import (
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

func Example_sync() {
	testTerm(func(term *vte.Terminal) {
		// Wait command completion as a GTK callback.
		term.Connect("child-exited", func() { printAndQuit(term) })

		// Start the command.
		_, e := term.ExecSync("", []string{"sh", "-c", bashCmd}, nil)
		if e != nil {
			fmt.Println(e)
		}
	})

	// Output:
	// closing in 3 sec
	// 2
	// 1
}

func Example_async() {
	testTerm(func(term *vte.Terminal) {
		term.Connect("child-exited", func() { printAndQuit(term) })
		term.ExecAsync(vtecommon.Cmd{
			Args:    []string{"sh", "-c", bashCmd},
			Timeout: -1,
			OnExec: func(pid int, e error) {
				if e != nil {
					fmt.Println(e)
				}
			},
		})
	})

	// Output:
	// closing in 3 sec
	// 2
	// 1
}

func Example_execCmd() {
	testTerm(func(term *vte.Terminal) {
		glib.IdleAdd(func() { // Wait gtk to be ready to start our command in the gtk loop.
			cmd := exec.Command("sh", "-c", bashCmd)
			cmd.Stdout = term // use the terminal as the command output (writer)
			cmd.Stderr = term

			e := cmd.Start()
			if e != nil {
				fmt.Println(e)
				return
			}

			// Use a go routine to release the gtk main loop and wait our command completion.
			go func() {
				cmd.Wait()

				glib.IdleAdd(func() { printAndQuit(term) }) // Resynced in the GTK loop to prevent thread crashes.
			}()
		})
	})

	// Output:
	// closing in 3 sec
	// 2
	// 1
}

func testTerm(callTest func(*vte.Terminal)) {
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

	callTest(term)

	// Start the main loop to run the test.
	gtk.Main()
}

func printAndQuit(term *vte.Terminal) {
	// Get the terminal text using the clipboard and print the result for the test.
	fmt.Print(term.GetText())

	gtk.MainQuit() // this release the main loop and finish the test.
}
