package js

import (
	"fmt"
	"github.com/Bios-Marcel/discordgo"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Bios-Marcel/cordless/scripting"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

// This declaration makes sure that JavaScriptEngine complies with the
// Engine interface.
var _ scripting.Engine = &JavaScriptEngine{}

// JavaScriptEngine stores scripting engine state
type JavaScriptEngine struct {
	scriptInstances []*ScriptInstance
	errorOutput     io.Writer
	// globalInstance is used for actions that don't require to be run on a
	// specific VM, but any VM. An example for this is converting a Go-struct
	// into a valid Otto-Value.
	globalInstance *otto.Otto
}

// ScriptInstance represents a usable and already loaded javascript. The
// callbacks are pre-evaluated and the instance can be locked as soon as any
// of the requested callbacks are available.
type ScriptInstance struct {
	vm   *otto.Otto
	lock sync.Mutex

	onMessageSend    *otto.Value
	onMessageReceive *otto.Value
	onMessageEdit    *otto.Value
	onMessageDelete  *otto.Value
}

// New instantiates a new scripting engine
func New() *JavaScriptEngine {
	return &JavaScriptEngine{}
}

// LoadScripts implements Engine. Each script gets a designated Otto-VM in
// order to avoid scripts modifying each others state by accident. All
// available callbacks get eagerly evaluated in the beginning. Locking of the
// instances when calling one of the callbacks only happens, if a callback
// actually exists.
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

		//Skip dot-folders and read non-dot-folders.
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
		_, err = vm.Run(file)
		if err != nil {
			return errors.Wrapf(err, "failed to run script '%s'", path)
		}

		instance := &ScriptInstance{
			vm:   vm,
			lock: sync.Mutex{},
		}

		onMessageSendJS, resolveError := vm.Get("onMessageSend")
		if !onMessageSendJS.IsUndefined() {
			instance.onMessageSend = &onMessageSendJS
		}
		if resolveError != nil {
			return errors.Wrap(resolveError, "error resolving function onMessageSend")
		}

		onMessageReceiveJS, resolveError := vm.Get("onMessageReceive")
		if !onMessageReceiveJS.IsUndefined() {
			instance.onMessageReceive = &onMessageReceiveJS
		}
		if resolveError != nil {
			return errors.Wrap(resolveError, "error resolving function onMessageReceive")
		}

		onMessageEditJS, resolveError := vm.Get("onMessageEdit")
		if !onMessageEditJS.IsUndefined() {
			instance.onMessageEdit = &onMessageEditJS
		}
		if resolveError != nil {
			return errors.Wrap(resolveError, "error resolving function onMessageEdit")
		}

		onMessageDeleteJS, resolveError := vm.Get("onMessageDelete")
		if !onMessageDeleteJS.IsUndefined() {
			instance.onMessageDelete = &onMessageDeleteJS
		}
		if resolveError != nil {
			return errors.Wrap(resolveError, "error resolving function onMessageDelete")
		}

		engine.scriptInstances = append(engine.scriptInstances, instance)
	}

	//Avoid unnecessarily creating an unused VM.
	if len(engine.scriptInstances) > 0 {
		engine.globalInstance = otto.New()
	}

	return nil
}

func (engine *JavaScriptEngine) SetErrorOutput(errorOutput io.Writer) {
	engine.errorOutput = errorOutput
}

// OnMessageSend implements Engine
func (engine *JavaScriptEngine) OnMessageSend(oldText string) (newText string) {
	newText = oldText
	for _, instance := range engine.scriptInstances {
		func() {
			if instance.onMessageSend != nil {
				defer instance.lock.Unlock()
				instance.lock.Lock()
				jsValue, jsError := instance.onMessageSend.Call(otto.NullValue(), newText)
				if jsError != nil {
					if engine.errorOutput != nil && !jsValue.IsUndefined() {
						fmt.Fprintf(engine.errorOutput, "Error occurred during execution of javascript: %s\n", jsError.Error())
					}
					//This script failed, go to next one
					return
				}
				newText = jsValue.String()
			}
		}()
	}

	return
}

// OnMessageReceive implements Engine
func (engine *JavaScriptEngine) OnMessageReceive(message *discordgo.Message) {
	if len(engine.scriptInstances) == 0 {
		return
	}

	messageToJS, toValueError := engine.globalInstance.ToValue(*message)
	if toValueError != nil {
		log.Printf("Error converting message to Otto value: %s\n", toValueError)
		return
	}

	for _, instance := range engine.scriptInstances {
		func() {
			if instance.onMessageReceive != nil {
				instance.lock.Lock()
				defer instance.lock.Unlock()

				_, callError := instance.onMessageReceive.Call(otto.NullValue(), messageToJS)
				if callError != nil {
					log.Printf("Error calling onMessageReceive: %s\n", callError)
				}
			}
		}()
	}
}

// OnMessageEdit implements Engine
func (engine *JavaScriptEngine) OnMessageEdit(message *discordgo.Message) {
	if len(engine.scriptInstances) == 0 {
		return
	}

	messageToJS, toValueError := engine.globalInstance.ToValue(*message)
	if toValueError != nil {
		log.Printf("Error converting message to Otto value: %s\n", toValueError)
		return
	}

	for _, instance := range engine.scriptInstances {
		func() {
			if instance.onMessageEdit != nil {
				instance.lock.Lock()
				defer instance.lock.Unlock()

				_, callError := instance.onMessageEdit.Call(otto.NullValue(), messageToJS)
				if callError != nil {
					log.Printf("Error calling onMessageEdit: %s\n", callError)
				}
			}
		}()
	}
}

// OnMessageDelete implements Engine
func (engine *JavaScriptEngine) OnMessageDelete(message *discordgo.Message) {
	if len(engine.scriptInstances) == 0 {
		return
	}

	messageToJS, toValueError := engine.globalInstance.ToValue(*message)
	if toValueError != nil {
		log.Printf("Error converting message to Otto value: %s\n", toValueError)
		return
	}

	for _, instance := range engine.scriptInstances {
		func() {
			if instance.onMessageDelete != nil {
				instance.lock.Lock()
				defer instance.lock.Unlock()
				_, callError := instance.onMessageDelete.Call(otto.NullValue(), messageToJS)
				if callError != nil {
					log.Printf("Error calling onMessageDelete: %s\n", callError)
				}
			}
		}()
	}
}

// SetTriggerNotificationFunction implements Engine
func (engine *JavaScriptEngine) SetTriggerNotificationFunction(function func(title, text string)) {
	triggerNotification := func(call otto.FunctionCall) otto.Value {
		title, argError := call.Argument(0).ToString()
		if argError != nil {
			log.Printf("Error invoking triggerNotification in JS engine: %s\n", argError)
			return otto.NullValue()
		}
		text, argError := call.Argument(1).ToString()
		if argError != nil {
			log.Printf("Error invoking triggerNotification in JS engine: %s\n", argError)
			return otto.NullValue()
		}
		function(title, text)
		return otto.NullValue()
	}
	engine.setFunctionOnVMs("triggerNotification", triggerNotification)
}

// SetPrintToConsoleFunction implements Engine
func (engine *JavaScriptEngine) SetPrintToConsoleFunction(function func(text string)) {
	printToConsole := func(call otto.FunctionCall) otto.Value {
		text, argError := call.Argument(0).ToString()
		if argError != nil {
			log.Printf("Error invoking printToConsole in JS engine: %s\n", argError)
		} else {
			function(text)
		}
		return otto.UndefinedValue()
	}
	engine.setFunctionOnVMs("printToConsole", printToConsole)
}

// SetPrintLineToConsoleFunction implements Engine
func (engine *JavaScriptEngine) SetPrintLineToConsoleFunction(function func(text string)) {
	printLineToConsole := func(call otto.FunctionCall) otto.Value {
		text, argError := call.Argument(0).ToString()
		if argError != nil {
			log.Printf("Error invoking printLineToConsole in JS engine: %s\n", argError)
		} else {
			function(text)
		}
		return otto.UndefinedValue()
	}
	engine.setFunctionOnVMs("printLineToConsole", printLineToConsole)
}

// SetGetCurrentGuildFunction implements Engine
func (engine *JavaScriptEngine) SetGetCurrentGuildFunction(function func() string) {
	getCurrentGuild := func(call otto.FunctionCall) otto.Value {
		guildID, callError := call.Otto.ToValue(function())
		if callError != nil {
			log.Printf("Error calling getCurrentGuild: %s\n", callError)
			return otto.NullValue()
		} else {
			return guildID
		}
	}
	engine.setFunctionOnVMs("getCurrentGuild", getCurrentGuild)
}

// SetGetCurrentChannelFunction implements Engine
func (engine *JavaScriptEngine) SetGetCurrentChannelFunction(function func() string) {
	getCurrentChannel := func(call otto.FunctionCall) otto.Value {
		guildID, callError := call.Otto.ToValue(function())
		if callError != nil {
			log.Printf("Error calling getCurrentChannel: %s\n", callError)
			return otto.NullValue()
		} else {
			return guildID
		}
	}
	engine.setFunctionOnVMs("getCurrentChannel", getCurrentChannel)
}

func (engine *JavaScriptEngine) setFunctionOnVMs(name string, function func(call otto.FunctionCall) otto.Value) {
	for _, instance := range engine.scriptInstances {
		setError := instance.vm.Set(name, function)
		if setError != nil {
			log.Printf("Error setting function %s: %s", name, setError)
		}
	}
}
