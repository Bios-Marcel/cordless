package vim

type VimMode int

const (
	// Disabled : Default mode. Vim disabled.
	Disabled int = iota
	// NormalMode : every key is interpreted as focus movement,
	// let it be from containers or text lines.
	NormalMode
	// InsertMode : keys inside text box are all sent to the input handler.
	// It also allows between-message movement inside chatview.
	InsertMode
	// Global movement with hjkl. Text selection in text box.
	VisualMode
)

// Vim stores information about vim-mode, such as
// current mode.
type Vim struct {
	// CurrentMode holds the integer value of the current vim mode.
	CurrentMode int
}

func NewVim(enabled bool) *Vim {
	if enabled {
		return &Vim{CurrentMode: NormalMode}
	} else {
		return &Vim{CurrentMode: Disabled}
	}
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

// SetNormal quickly switches to normal mode.
func (v *Vim) SetNormal() {
	v.CurrentMode = NormalMode
}

// SetInsert quickly switches to insert mode.
func (v *Vim) SetInsert() {
	v.CurrentMode = InsertMode
}

// SetVisual quickly switches to visual mode.
func (v *Vim) SetVisual() {
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
