// Package vte is a cgo binding for Vte. Supports version 2.91 (0.40) and later.
//
// This package provides the Vte terminal without any GTK dependency.
//
// https://developer.gnome.org/vte/0.40/VteTerminal.html
//
// https://developer.gnome.org/vte/unstable/VteTerminal.html
//
package vte

/*
#include <stdlib.h>
#include <vte/vte.h>

// Go exported func redeclarations.
extern void onAsyncOnExec (VteTerminal *terminal, GPid pid, GError *error, gpointer callback);




static inline char** make_strings(int count) {
	return (char**)malloc(sizeof(char*) * count);
}

static inline void set_string(char** strings, int n, char* str) {
	strings[n] = str;
}


static gpointer      uintToGpointer (uint i)      { return GUINT_TO_POINTER(i); }
static uint          gpointerToUint (gpointer i)  { return GPOINTER_TO_UINT(i); }
static VteTerminal * toVteTerminal (void *p)      { return (VTE_TERMINAL(p)); }

*/
// #cgo pkg-config: vte-2.91
import "C"

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// Format defines the format the selection should be copied to the clipboard in.
//
type Format int32

// Text formats for clipboard copy.
const (
	FormatText Format = C.VTE_FORMAT_TEXT // Export as plain text
	FormatHTML Format = C.VTE_FORMAT_HTML // Export as HTML formatted text
)

// Colors palette names.
//
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
//
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
//
type Terminal struct {
	ptr *C.VteTerminal
}

// NewTerminal is a wrapper around vte_terminal_new().
//
func NewTerminal() *Terminal {
	c := C.vte_terminal_new()
	if c == nil {
		return nil
	}
	return &Terminal{C.toVteTerminal(unsafe.Pointer(c))}
}

// Native returns a pointer to the underlying VteTerminal.
//
func (v *Terminal) Native() *C.VteTerminal {
	return v.ptr
}

// Feed interprets data as if it were data received from a child process.
//
func (v *Terminal) Feed(m string) {
	c := C.CString(m)
	defer C.free(unsafe.Pointer(c))
	C.vte_terminal_feed(v.Native(), C.CString(m), -1)
}

// FeedChild sends a block of UTF-8 text to the child as if it were entered by
// the user at the keyboard.
//
func (v *Terminal) FeedChild(m string) {
	c := C.CString(m)
	defer C.free(unsafe.Pointer(c))
	C.vte_terminal_feed_child(v.Native(), C.CString(m), -1)
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
func GetUserShell() string {
	c := C.vte_get_user_shell()
	defer C.free(unsafe.Pointer(c))
	return C.GoString(c)
}

// Cmd represents an external command being prepared or run.
//
type Cmd struct {
	Dir     string            // Dir specifies the working directory of the command.
	Args    []string          // Args holds command line arguments, including the command as Args[0].
	Timeout int               // Timeout specifies the time allowed in ms, or -1 to wait indefinitely.
	Env     map[string]string // Env specifies the environment of the process.
	OnExec  func(int, error)  // OnExec is called when the process is started (or failed to).
	// Cancellable *C.GCancellable // TODO.
}

// NewCmd creates a new command to run async in the terminal.
//
func (v *Terminal) NewCmd(args ...string) Cmd {
	return Cmd{Args: args, Timeout: -1}
}

// ExecAsync starts the given command in the terminal.
//
func (v *Terminal) ExecAsync(cmd Cmd) {

	var ccwd *C.char
	if cmd.Dir != "" {
		ccwd = C.CString(cmd.Dir)
		defer C.free(unsafe.Pointer(ccwd))
	}

	cargs := C.make_strings(C.int(len(cmd.Args)) + 1)
	for i, j := range cmd.Args {
		ptr := C.CString(j)
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cargs, C.int(i), ptr)
	}
	C.set_string(cargs, C.int(len(cmd.Args)), nil) // null terminated list.

	cenv := C.make_strings(C.int(len(cmd.Env)) + 1)
	i := 0
	for k, v := range cmd.Env {
		ptr := C.CString(fmt.Sprintf("%s=%s", k, v))
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cenv, C.int(i), ptr)
		i++
	}
	C.set_string(cenv, C.int(len(cmd.Env)), nil) // null terminated list.

	var ccallID C.gpointer
	if cmd.OnExec != nil {
		callID := assignCallID(cmd)
		if callID == 0 {
			cmd.OnExec(0, errors.New("spawn sync is unable to store the callback for: "+strings.Join(cmd.Args, " ")))
			return
		}
		ccallID = C.uintToGpointer(C.uint(callID))
	}

	C.vte_terminal_spawn_async(v.Native(),
		C.VTE_PTY_DEFAULT, // VtePtyFlags
		ccwd,              // const char *working_directory
		cargs,             // char **argv
		cenv,              // char **envv
		C.G_SPAWN_SEARCH_PATH, // GSpawnFlags
		nil,                // GSpawnChildSetupFunc
		nil,                // gpointer child_setup_data
		nil,                // GDestroyNotify for child_setup_data_destroy
		C.int(cmd.Timeout), // int
		nil,                // GCancellable
		C.VteTerminalSpawnAsyncCallback(C.onAsyncOnExec), // VteTerminalSpawnAsyncCallback
		ccallID, // gpointer user_data
	)
}

var asyncCallIDs = make(map[uint]Cmd)
var asyncCallMU = sync.Mutex{}

func assignCallID(cmd Cmd) uint {
	callID := uint(1)
	asyncCallMU.Lock()
	defer asyncCallMU.Unlock()
	for callID != 0 {
		_, isset := asyncCallIDs[callID]
		if !isset {
			asyncCallIDs[callID] = cmd
			return callID
		}
		callID++
	}
	return 0
}

//export onAsyncOnExec
//
// called when ExecAsync process is started or failed.
//
func onAsyncOnExec(terminal *C.VteTerminal, cpid C.GPid, cerr *C.GError, ccallID C.gpointer) {
	callID := uint(C.gpointerToUint(ccallID))
	if callID == 0 {
		return
	}

	asyncCallMU.Lock()
	cmd, ok := asyncCallIDs[callID]
	if !ok {
		fmt.Printf("onAsyncOnExec can't find callback ID:%d\n", callID)
		asyncCallMU.Unlock()
		return
	}
	delete(asyncCallIDs, callID)
	asyncCallMU.Unlock()

	var e error
	if cerr != nil {
		e = errors.New(C.GoString((*C.char)(cerr.message)))
	}
	cmd.OnExec(int(cpid), e)
}

// ExecSync starts the given command in the terminal. Deprecated since 0.48.
// It's a wrapper around vte_terminal_spawn_sync.
//
func (v *Terminal) ExecSync(cwd string, args []string, env map[string]string) (int, error) {

	var ccwd *C.char
	if cwd != "" {
		ccwd = C.CString(cwd)
		defer C.free(unsafe.Pointer(ccwd))
	}

	cargs := C.make_strings(C.int(len(args)) + 1)
	for i, j := range args {
		ptr := C.CString(j)
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cargs, C.int(i), ptr)
	}
	C.set_string(cargs, C.int(len(args)), nil) // null terminated list.

	cenv := C.make_strings(C.int(len(env)) + 1)
	i := 0
	for k, v := range env {
		ptr := C.CString(fmt.Sprintf("%s=%s", k, v))
		defer C.free(unsafe.Pointer(ptr))
		C.set_string(cenv, C.int(i), ptr)
		i++
	}
	C.set_string(cenv, C.int(len(env)), nil) // null terminated list.

	var cerr *C.GError
	var cpid C.GPid

	C.vte_terminal_spawn_sync(v.Native(),
		C.VTE_PTY_DEFAULT, // VtePtyFlags
		ccwd,              // const char *working_directory
		cargs,             // char **argv
		cenv,              // char **envv
		C.G_SPAWN_SEARCH_PATH, // GSpawnFlags
		nil,   // GSpawnChildSetupFunc
		nil,   // gpointer child_setup_data
		&cpid, // GPid *child_pid
		nil,   // GCancellable *cancellable
		&cerr, // GError **error
	)
	if cerr != nil {
		defer C.g_error_free(cerr)
		return 0, errors.New(C.GoString((*C.char)(cerr.message)))
	}

	return int(cpid), nil
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

// SetFontScale sets the terminal's font scale to scale.
//
func (v *Terminal) SetFontScale(scale float64) {
	C.vte_terminal_set_font_scale(v.Native(), C.gdouble(scale))
}

// GetFontScale returns the terminal's font scale.
//
func (v *Terminal) GetFontScale() float64 {
	return float64(C.vte_terminal_get_font_scale(v.Native()))
}

// SetScrollbackLines sets the length of the scrollback buffer used by the
// terminal. The size of the scrollback buffer will be set to the larger of this
// value and the number of visible rows the widget can display, so 0 can safely
// be used to disable scrollback.
//
// A negative value means "infinite scrollback".
//
// Note that this setting only affects the normal screen buffer.
// No scrollback is allowed on the alternate screen buffer.
//
func (v *Terminal) SetScrollbackLines(val int32) {
	C.vte_terminal_set_scrollback_lines(v.Native(), C.glong(val))
}

// GetCursorPosition reads the location of the insertion cursor and returns it.
// The row coordinate is absolute.
//
func (v *Terminal) GetCursorPosition() (int32, int32) {
	var column, row C.glong
	C.vte_terminal_get_cursor_position(v.Native(), &column, &row)
	return int32(column), int32(row)
}

// GetText extracts a view of the visible part of the terminal.
//
func (v *Terminal) GetText() string {
	data := C.vte_terminal_get_text(v.Native(),
		nil,
		nil,
		nil)
	return C.GoString(data)
}

// GetTextRange Extracts a view of the visible part of the terminal.
//
func (v *Terminal) GetTextRange(startRow int32, startCol int32, endRow int32, endCol int32) string {
	data := C.vte_terminal_get_text_range(v.Native(),
		C.glong(startRow),
		C.glong(startCol),
		C.glong(endRow),
		C.glong(endCol),
		nil,
		nil,
		nil)
	return C.GoString(data)
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

// GetWindowTitle get the terminal window title.
//
func (v *Terminal) GetWindowTitle() string {
    title := C.vte_terminal_get_window_title(v.Native())
    return C.GoString(title)
}

// SetEncoding sets the terminal encoding data.
//
func (v *Terminal) SetEncoding(s string) {
    cstr := C.CString(s)
    C.vte_terminal_set_encoding(v.Native(), cstr, nil)
}

// GetEncoding get the name of the encoding in which the terminal
// expects data to be encoded.
// 
func (v *Terminal) GetEncoding() string {
    encoding := C.vte_terminal_get_encoding(v.Native())
    return C.GoString(encoding)
}

// SelectAll selects all text within the terminal (including the scrollback buffer).
//
func (v *Terminal) SelectAll() {
	C.vte_terminal_select_all(v.Native())
}

// UnSelectAll clears the current selection.
//
func (v *Terminal) UnSelectAll() {
	C.vte_terminal_unselect_all(v.Native())
}

// CopyClipboard places the selected text in the terminal in the
// GDK_SELECTION_CLIPBOARD selection.
//
// Deprecated since 0.50. Use CopyClipboardFormat with FormatText instead.
//
func (v *Terminal) CopyClipboard() {
	C.vte_terminal_copy_clipboard(v.Native())
}

// CopyClipboardFormat places the selected text in the terminal in the
// GDK_SELECTION_CLIPBOARD selection in the form specified by format.
//
func (v *Terminal) CopyClipboardFormat(format Format) {
	C.vte_terminal_copy_clipboard_format(v.Native(), C.VteFormat(format))
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

// WatchChild watches child_pid.
//
// When the process exists, the “child-exited” signal will be called with the child's exit status.
//
func (v *Terminal) WatchChild(pid int) {
	C.vte_terminal_watch_child(v.Native(), C.GPid(pid))
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
