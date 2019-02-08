package scripting

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

var _ Engine = &JavaScriptEngine{}

// JavaScriptEngine stores scripting engine state
type JavaScriptEngine struct {
	vms []*otto.Otto
}

// New instantiates a new scripting engine
func New() (engine *JavaScriptEngine) {
	engine = &JavaScriptEngine{
		vms: make([]*otto.Otto, 0),
	}

	return
}

// LoadScripts implements Engine
func (engine *JavaScriptEngine) LoadScripts(dirname string) (err error) {
	_, statError := os.Stat(dirname)
	if os.IsNotExist(statError) {
		return nil
	} else if statError != nil {
		return errors.Wrapf(statError, "Error loading scripts '%s'", statError.Error())
	}

	return engine.readScriptsRecursively(dirname)
}

func (engine *JavaScriptEngine) readScriptsRecursively(dirname string) error {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}

	for _, file := range files {
		path := filepath.Join(dirname, file.Name())

		//Skip dotfolders and read non-dotfolders.
		if file.IsDir() {
			if !strings.HasPrefix(file.Name(), ".") {
				readError := engine.readScriptsRecursively(path)
				if readError != nil {
					return readError
				}
			}

			continue
		}

		//Only javascript files
		if !strings.HasSuffix(file.Name(), ".js") {
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, path)
		}

		vm := otto.New()
		engine.vms = append(engine.vms, vm)
		_, err = vm.Run(file)
		if err != nil {
			return errors.Wrapf(err, "failed to run script '%s'", path)
		}
	}

	return nil
}

// OnMessageSend implements Engine
func (engine *JavaScriptEngine) OnMessageSend(oldText string) (newText string) {
	newText = oldText
	for _, vm := range engine.vms {
		jsValue, jsError := vm.Run(fmt.Sprintf("onMessageSend(\"%s\")", escapeNewlines(newText)))
		if jsError != nil {
			//This script failed, go to next one
			continue
		}
		newText = jsValue.String()
	}

	return
}

func escapeNewlines(parameter string) string {
	return strings.Replace(parameter, "\n", "\\n", -1)
}
