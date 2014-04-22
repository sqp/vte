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
// #cgo pkg-config: vte-2.90
import "C"

import (
	"unsafe"
)

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
	White:        "#E7E7E7",
	WhiteLight:   "#E7E7E&",
}

type Terminal struct {
	term *C.VteTerminal
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
	return v.term
}

func (v *Terminal) Feed(m string) {
	c := C.CString(m)
	defer C.free(unsafe.Pointer(c))
	C.vte_terminal_feed(v.Native(), C.CString(m), -1)
}

func (v *Terminal) Fork(args []string) {
	cargs := C.make_strings(C.int(len(args)))
	for i, j := range args {
		ptr := C.CString(j)
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cargs, C.int(i), ptr)
	}
	C.vte_terminal_fork_command_full(v.Native(),
		C.VTE_PTY_DEFAULT,
		nil,
		cargs,
		nil,
		C.G_SPAWN_SEARCH_PATH,
		nil,
		nil,
		nil, nil)
}

func (v *Terminal) SetFontFromString(font string) {
	cstr := C.CString(font)
	defer C.free(unsafe.Pointer(cstr))
	// C.vte_terminal_set_emulation(s.Native(), cstr)
	C.vte_terminal_set_font_from_string(v.Native(), cstr)
}

type Palette struct {
}

func (v *Terminal) SetBgColor(s string) {
	// C.vte_terminal_set_color_background(v.Native(), getColor(s))
}

func (v *Terminal) SetFgColor(s string) {
	// C.vte_terminal_set_color_foreground(v.Native(), getColor(s))
}

func (v *Terminal) SetColors(pal map[int]string) {
	colors := new([16]C.GdkColor)
	for i := 0; i < len(colors); i++ {
		C.gdk_color_parse((*C.gchar)(C.CString(pal[i])), &colors[i])
	}
	C.vte_terminal_set_colors(
		v.Native(),
		nil, nil,
		(*C.GdkColor)(unsafe.Pointer(colors)),
		16)
}

// func getColor(s string) *C.GdkColor {
// 	c := gdk.Color(s).Color
// 	return (*C.GdkColor)(unsafe.Pointer(&c))
// }
