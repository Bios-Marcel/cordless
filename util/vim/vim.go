package vim

const (
	// NormalMode : every key is interpreted as focus movement,
	// let it be from containers or text lines.
	NormalMode int = iota	// 0
	// InsertMode : every key is sent to the program as normal,
	// such as input boxes.
	InsertMode				// 1
	// TODO
	VisualMode				// 2
)

// Vim stores information about vim-mode, such as
// current mode.
type Vim struct {
	CurrentMode int
}

// SetMode sets new vim mode. If provided mode falls out
// the defined range 0-2 it will fall back to normal mode.
func (v *Vim) SetMode(mode int) {
	if mode < 0 || mode > 2 {
		// fallback
		v.CurrentMode = NormalMode
		return
	}
	v.CurrentMode = mode
}

// CurrentModeString returns a stringified version
// of the current mode. (Useful for visual feedback to user)
func (v *Vim) CurrentModeString() string {
	switch v.CurrentMode {
	case NormalMode:
		return "Normal"
	case InsertMode:
		return "Insert"
	case VisualMode:
		return "Visual"
	default:
		return "Err"
	}
}
