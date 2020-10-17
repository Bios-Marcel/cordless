package windowman

type DialogCloser func() error
type Dialog interface {
	Window
	Open(close DialogCloser) error
}
