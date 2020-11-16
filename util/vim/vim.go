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

	// Default: disabled
	Disabled = -1
)

// Vim stores information about vim-mode, such as
// current mode.
type Vim struct {
	// CurrentMode holds the integer value of the current vim mode.
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

// Quick mode switch

// Normal quickly switches to normal mode.
func (v *Vim) Normal() {
	v.CurrentMode = NormalMode
}

// Insert quickly switches to insert mode.
func (v *Vim) Insert() {
	v.CurrentMode = InsertMode
}

// Visual quickly switches to visual mode.
func (v *Vim) Visual() {
	v.CurrentMode = VisualMode
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
	case Disabled:
		return "Disabled"
	// Should not be needed, but better to always add a default case.
	default:
		return "Disabled"
	}
}

func (v *Vim) EnabledString() string {
	if v.CurrentMode == Disabled {
		return "Disabled"
	} else {
		return "Enabled"
	}
}
