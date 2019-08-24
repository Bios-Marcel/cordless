// +build !linux,!darwin!,!netbsd,!openbsd

package fdwscreen

type fdSwapper struct{}

func newFdSwapper() (*fdSwapper, error) {
	return nil, nil
}

func (s *fdSwapper) InitSwap() {
}

func (s *fdSwapper) FiniSwap() {
}
