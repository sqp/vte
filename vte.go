// Package vte is a binding for Vte. Supports version 2.91 (0.40) and later.
//
// This package provides the Vte terminal without any GTK dependency.
//
// https://developer.gnome.org/vte/0.40/VteTerminal.html
package vte

/*
#include <stdlib.h>
#include <vte/vte.h>

static inline char** make_strings(int count) {
	return (char**)malloc(sizeof(char*) * count);
}

static inline void set_string(char** strings, int n, char* str) {
	strings[n] = str;
}

static VteTerminal * toVteTerminal (void *p) { return (VTE_TERMINAL(p)); }

*/
// #cgo pkg-config: vte-2.91
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

// Colors palette names.
const (
	Black = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BlackLight
	RedLight
	GreenLight
	YellowLight
	BlueLight
	MagentaLight
	CyanLight
	WhiteLight
)

// MikePal defines a color palette example.
var MikePal = map[int]string{
	Black:        "#000000",
	BlackLight:   "#252525",
	Red:          "#803232",
	RedLight:     "#982B2B",
	Green:        "#85A136",
	GreenLight:   "#85A136",
	Yellow:       "#AA9943",
	YellowLight:  "#EFEF60",
	Blue:         "#324C80",
	BlueLight:    "#4186BE",
	Magenta:      "#706C9A",
	MagentaLight: "#826AB1",
	Cyan:         "#92B19E",
	CyanLight:    "#A1CDCD",
	White:        "#FFFFFF", // E7E7E7
	WhiteLight:   "#E7E7E&",
}

// Terminal is a representation of Vte's VteTerminal.
type Terminal struct {
	ptr *C.VteTerminal
}

// NewTerminal is a wrapper around vte_terminal_new().
func NewTerminal() *Terminal {
	c := C.vte_terminal_new()
	if c == nil {
		return nil
	}
	return &Terminal{C.toVteTerminal(unsafe.Pointer(c))}
}

// Native() returns a pointer to the underlying VteTerminal.
func (v *Terminal) Native() *C.VteTerminal {
	return v.ptr
}

func (v *Terminal) Feed(m string) {
	c := C.CString(m)
	defer C.free(unsafe.Pointer(c))
	C.vte_terminal_feed(v.Native(), C.CString(m), -1)
}

// Write forward the stream to the connected logger.
//
func (v *Terminal) Write(p []byte) (n int, err error) {
	v.Feed(strings.Replace(string(p), "\n", "\r\n", -1))

	return len(p), nil
}

// GetUserShell gets the user's shell.
// If empty, the system default (usually "/bin/sh") should be used).
//
func (v *Terminal) GetUserShell() string {
	c := C.vte_get_user_shell()
	defer C.free(unsafe.Pointer(c))
	return C.GoString(c)
}

// Fork starts the given command in the terminal.
// It's a simple wrapper around vte_terminal_spawn_sync which could be improved.
//
func (v *Terminal) Fork(args ...string) error {
	cargs := C.make_strings(C.int(len(args)) + 1)
	for i, j := range args {
		ptr := C.CString(j)
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cargs, C.int(i), ptr)
	}
	C.set_string(cargs, C.int(len(args)), nil) // null terminated list.

	var cerr *C.GError = nil

	C.vte_terminal_spawn_sync(v.Native(),
		C.VTE_PTY_DEFAULT, // VtePtyFlags
		nil,               // const char *working_directory
		cargs,             // char **argv
		nil,               // char **envv
		C.G_SPAWN_SEARCH_PATH, // GSpawnFlags
		nil,   // GSpawnChildSetupFunc
		nil,   // gpointer child_setup_data
		nil,   // GPid *child_pid
		nil,   // GCancellable *cancellable
		&cerr, // GError **error
	)
	if cerr != nil {
		defer C.g_error_free(cerr)
		return errors.New(C.GoString((*C.char)(cerr.message)))
	}
	return nil
}

// SetBgColorFromString sets the background color for text which does not have a
// specific background color assigned. Only has effect when no background image
// is set and when the terminal is not transparent.
// The color string is used as input.
//
func (v *Terminal) SetBgColorFromString(s string) {
	color := new(C.GdkRGBA)
	parseColor(s, color)
	C.vte_terminal_set_color_background(v.Native(), color)
}

// SetFgColorFromString sets the foreground color used to draw normal text using
// the color string.
//
func (v *Terminal) SetFgColorFromString(s string) {
	color := new(C.GdkRGBA)
	parseColor(s, color)
	C.vte_terminal_set_color_foreground(v.Native(), color)
}

// SetFontFromString sets the font used for rendering all text displayed by the
// terminal, overriding any fonts set using gtk_widget_modify_font().
// The terminal will immediately attempt to load the desired font, retrieve its
// metrics, and attempt to resize itself to keep the same number of rows and
// columns. The font scale is applied to the specified font.
// The font string is used as input.
//
func (v *Terminal) SetFontFromString(font string) {
	cstr := C.CString(font)
	defer C.free(unsafe.Pointer(cstr))
	c := C.pango_font_description_from_string(cstr)
	C.vte_terminal_set_font(v.Native(), c)
}

// SetColorsFromStrings sets the terminal color palette.
// A color strings map filled with the index starting at 0 is used as input.
//
func (v *Terminal) SetColorsFromStrings(pal map[int]string) error {
	var c *C.GdkRGBA
	switch len(pal) {
	case 16:
		colors := new([16]C.GdkRGBA)
		for i := 0; i < len(colors); i++ {
			C.gdk_rgba_parse(&colors[i], (*C.gchar)(C.CString(pal[i])))
			parseColor(pal[i], &colors[i])
		}
		c = (*C.GdkRGBA)(unsafe.Pointer(colors))

	// case 0, 8, 232, 256: // should be fine to do the same.

	default:
		return errors.New("SetColorsString: bad size, need 16 color strings")
	}
	C.vte_terminal_set_colors(
		v.Native(),
		nil, nil,
		c,
		C.gsize(len(pal)))
	return nil
}

// HasSelection checks if the terminal currently contains selected text.
// Note that this is different from determining if the terminal is the owner of
// any GtkClipboard items.
//
func (v *Terminal) HasSelection() bool {
	c := C.vte_terminal_get_has_selection(v.Native())
	if c == 0 {
		return false
	}
	return true
}

// SelectAll selects all text within the terminal (including the scrollback buffer).
//
func (v *Terminal) SelectAll() {
	C.vte_terminal_select_all(v.Native())
}

// CopyClipboard places the selected text in the terminal in the
// GDK_SELECTION_CLIPBOARD selection.
//
func (v *Terminal) CopyClipboard() {
	C.vte_terminal_copy_clipboard(v.Native())
}

// PasteClipboard sends the contents of the GDK_SELECTION_CLIPBOARD selection to
// the terminal's child. If necessary, the data is converted from UTF-8 to the
// terminal's current encoding.
//
func (v *Terminal) PasteClipboard() {
	C.vte_terminal_paste_clipboard(v.Native())
}

// CopyPrimary places the selected text in the terminal in the
// GDK_SELECTION_PRIMARY selection.
//
func (v *Terminal) CopyPrimary() {
	C.vte_terminal_copy_primary(v.Native())
}

// PastePrimary sends the contents of the GDK_SELECTION_PRIMARY selection to the
// terminal's child. If necessary, the data is converted from UTF-8 to the
// terminal's current encoding.
//
func (v *Terminal) PastePrimary() {
	C.vte_terminal_paste_primary(v.Native())
}

// Reset resets as much of the terminal's internal state as possible,
// discarding any unprocessed input data, resetting character attributes,
// cursor state, national character set state, status line,
// terminal modes (insert/delete), selection state, and encoding.
//
//   clearTabstops: whether to reset tabstops.
//   clearHistory:  whether to empty the terminal's scrollback buffer.
//
func (v *Terminal) Reset(clearTabstops, clearHistory bool) {
	C.vte_terminal_reset(v.Native(), cbool(clearTabstops), cbool(clearHistory))
}

func parseColor(s string, color *C.GdkRGBA) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C.gdk_rgba_parse(color, (*C.gchar)(cstr))
}

func cbool(b bool) C.gboolean {
	if b {
		return 1
	}
	return 0
}
