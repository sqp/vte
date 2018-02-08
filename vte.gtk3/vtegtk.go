// Package vte is a cgo binding for Vte. Supports version 2.91 (0.40) and later.
//
// This package provides the Vte terminal wrapped as a gotk3 widget, and using
// this library ressources.
//
// https://developer.gnome.org/vte/0.40/VteTerminal.html
package vte

// #include <vte/vte.h>
// #cgo pkg-config: vte-2.91
import "C"

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/sqp/vte"

	"errors"
	"runtime"
	"unsafe"
)

// Terminal is a representation of Vte's VteTerminal.
//
type Terminal struct {
	gtk.Widget
	vte.Terminal
}

// NewTerminal creates a new terminal widget.
//
func NewTerminal() *Terminal {
	c := vte.NewTerminal()
	if c == nil {
		return nil
	}

	obj := &glib.Object{glib.ToGObject(unsafe.Pointer(c.Native()))}
	obj.RefSink()
	runtime.SetFinalizer(obj, (*glib.Object).Unref)

	return wrapTerminal(obj, c)
}

// NewTerminalWindow creates a new terminal widget packed in a dedicated window.
//
func NewTerminalWindow() (*Terminal, *gtk.Window, error) {
	window, e := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if e != nil {
		return nil, nil, e
	}

	terminal := NewTerminal()
	if terminal == nil {
		return nil, nil, errors.New("create terminal failed")
	}

	// Packing.
	swin, _ := gtk.ScrolledWindowNew(nil, nil)
	swin.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	swin.Add(terminal)
	window.Add(swin)
	window.ShowAll()

	return terminal, window, nil
}

func (v *Terminal) termNative() *C.VteTerminal {
	return (*C.VteTerminal)(unsafe.Pointer(v.Terminal.Native()))
}

func wrapTerminal(obj *glib.Object, term *vte.Terminal) *Terminal {
	return &Terminal{gtk.Widget{glib.InitiallyUnowned{obj}}, *term}
}

// SetBgColor sets the background color for text which does not have a specific
// background color assigned. Only has effect when no background image is set
// and when the terminal is not transparent.
// The gdk RGBA is used as input.
//
func (v *Terminal) SetBgColor(color *gdk.RGBA) {
	C.vte_terminal_set_color_background(v.termNative(), (*C.GdkRGBA)(unsafe.Pointer(color.Native())))
}

// SetFgColor sets the foreground color used to draw normal text.
// The gdk RGBA is used as input.
//
func (v *Terminal) SetFgColor(color *gdk.RGBA) {
	C.vte_terminal_set_color_foreground(v.termNative(), (*C.GdkRGBA)(unsafe.Pointer(color.Native())))
}

// SetFontFromString sets the font used for rendering all text displayed by the
// terminal, overriding any fonts set using widget.ModifyFont().
// The terminal will immediately attempt to load the desired font, retrieve its
// metrics, and attempt to resize itself to keep the same number of rows and
// columns. The font scale is applied to the specified font.
// The pango FontDescription is used as input.
//
func (v *Terminal) SetFont(font *pango.FontDescription) {
	C.vte_terminal_set_font(v.termNative(), (*C.PangoFontDescription)(unsafe.Pointer(font.Native())))
}

func (v *Terminal) SetFontScale(scale float64) {
	C.vte_terminal_set_font_scale(v.termNative(), C.gdouble(scale))
}

func (v *Terminal) GetFontScale() float64 {
	return float64(C.vte_terminal_get_font_scale(v.termNative()))
}

func (v *Terminal) SetScrollbackLines(val int32) {
	C.vte_terminal_set_scrollback_lines(v.termNative(), C.glong(val))
}
