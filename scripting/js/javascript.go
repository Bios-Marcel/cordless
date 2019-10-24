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

	"github.com/Bios-Marcel/cordless/scripting"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

// This decleration makes sure that JavaScriptEngine complies with the
// Engine interface.
var _ scripting.Engine = &JavaScriptEngine{}

// JavaScriptEngine stores scripting engine state
type JavaScriptEngine struct {
	vms         []*otto.Otto
	errorOutput io.Writer
}

// New instantiates a new scripting engine
func New() *JavaScriptEngine {
	return &JavaScriptEngine{}
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

func (engine *JavaScriptEngine) SetErrorOutput(errorOutput io.Writer) {
	engine.errorOutput = errorOutput
}

func escapeNewlines(parameter string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"\n", "\\n",
		"\"", "\\\"").
		Replace(parameter)
}

// OnMessageSend implements Engine
func (engine *JavaScriptEngine) OnMessageSend(oldText string) (newText string) {
	newText = oldText
	for _, vm := range engine.vms {
		jsValue, jsError := vm.Run(fmt.Sprintf("onMessageSend(\"%s\")", escapeNewlines(newText)))
		if jsError != nil {
			if engine.errorOutput != nil && !jsValue.IsUndefined() {
				fmt.Fprintf(engine.errorOutput, "Error occurred during execution of javascript: %s\n", jsError.Error())
			}
			//This script failed, go to next one
			continue
		}
		newText = jsValue.String()
	}

	return
}

// OnMessageRead implements Engine
func (engine *JavaScriptEngine) OnMessageReceive(message *discordgo.Message) {
	for _, vm := range engine.vms {
		onMessageReceiveJS, resolveError := vm.Get("onMessageReceive")
		if onMessageReceiveJS.IsUndefined() {
			continue
		}
		if resolveError != nil {
			log.Printf("Error resolving function onMessageReceive: %s\n", resolveError)
			continue
		}

		messageToJS, toValueError := vm.ToValue(*message)
		if toValueError != nil {
			log.Printf("Error converting message to Otto value: %s\n", toValueError)
		} else {
			_, callError := onMessageReceiveJS.Call(otto.NullValue(), messageToJS)
			if callError != nil {
				log.Printf("Error calling onMessageReceive: %s\n", callError)
			}
		}
	}
}

// OnMessageDelete implements Engine
func (engine *JavaScriptEngine) OnMessageDelete(message *discordgo.Message) {
	for _, vm := range engine.vms {
		onMessageDeleteJS, resolveError := vm.Get("onMessageDelete")
		if onMessageDeleteJS.IsUndefined() {
			continue
		}
		if resolveError != nil {
			log.Printf("Error resolving function onMessageDelete: %s\n", resolveError)
			continue
		}

		messageToJS, toValueError := vm.ToValue(*message)
		if toValueError != nil {
			log.Printf("Error converting message to Otto value: %s\n", toValueError)
		} else {
			_, callError := onMessageDeleteJS.Call(otto.NullValue(), messageToJS)
			if callError != nil {
				log.Printf("Error calling onMessageDelete: %s\n", callError)
			}
		}
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
	for _, vm := range engine.vms {
		setError := vm.Set(name, function)
		if setError != nil {
			log.Printf("Error setting function %s: %s", name, setError)
		}
	}
}
