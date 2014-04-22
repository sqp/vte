package vte

import (
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"github.com/sqp/vte"

	"runtime"
	"unsafe"
)

type Terminal struct {
	gtk.Widget
	vte.Terminal
}

// NewTerminal is a wrapper around vte_terminal_new().
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

func wrapTerminal(obj *glib.Object, term *vte.Terminal) *Terminal {
	return &Terminal{gtk.Widget{glib.InitiallyUnowned{obj}}, *term}
}
