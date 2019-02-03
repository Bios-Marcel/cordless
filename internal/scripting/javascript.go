package scripting

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

var _ Engine = &JavaScriptEngine{}

// JavaScriptEngine stores scripting engine state
type JavaScriptEngine struct {
	vm *otto.Otto
}

// New instantiates a new scripting engine
func New() (e *JavaScriptEngine) {
	e = &JavaScriptEngine{
		vm: otto.New(),
	}

	return
}

// LoadScripts implements Engine
func (e *JavaScriptEngine) LoadScripts(dirname string) (err error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(dirname, file.Name())
		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, path)
		}
		_, err = e.vm.Run(f)
		if err != nil {
			return errors.Wrap(err, "failed to run script")
		}
	}

	return
}

// OnMessage implements Engine
func (e *JavaScriptEngine) OnMessage(oldText string) (newText string) {
	v, err := e.vm.Run(fmt.Sprintf(`onMessage("%s")`, oldText))
	if err != nil {
		return oldText
	}
	return v.String()
}
