package server

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type PathParameters interface {
	Get(param string) (string, *RequestError)
	AsMap() map[string]string
	// TODO Add GetAsInt, etc. Will need to distinguish between missing and invalid
}

type parameterMap map[string]string

func (m parameterMap) Get(param string) (string, *RequestError) {
	// TODO Return 400 is string is missing
	return m[param], nil
}

func (m parameterMap) AsMap() map[string]string {
	return m
}

// TODO Should we have a type with no PathParameters?
type GetHandler func(params PathParameters) (interface{}, *RequestError)

// TODO Fix naming i.e. handlerMutex.mutex
type handlerMutex struct {
	mutex    sync.RWMutex
	handlers map[string]mutexEntry
}

type mutexEntry struct {
	typeName string
	parameters []string
	handler GetHandler
}

func newHandlerMutex() *handlerMutex { return &handlerMutex{handlers: make(map[string]mutexEntry)} }

func (mutex *handlerMutex) registerHandler(typeName string, parameters []string, handler GetHandler) {
	mutex.mutex.Lock()
	defer mutex.mutex.Unlock()

	mutex.handlers[typeName] = mutexEntry{typeName: typeName, parameters: parameters, handler: handler }
}

func (mutex *handlerMutex) getHandler(typeName string) (handler GetHandler, parameters []string) {
	mutex.mutex.RLock()
	defer mutex.mutex.RUnlock()

	// TODO Handle missing map entry
	entry := mutex.handlers[typeName]
	return entry.handler, entry.parameters
}

func ClearHandlers() {
	// FIXME This is not threadsafe, but is just used for tests ATM
	defaultHandlerMutex = newHandlerMutex()
}

var defaultHandlerMutex = newHandlerMutex()

func getInterfaceTypeName(i interface{}) (t reflect.Type, name string) {
	t = reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		pointerToValue := reflect.ValueOf(i)
		valueAtPointer := reflect.Indirect(pointerToValue)
		t = valueAtPointer.Type()
	}
	name = t.Name()
	return
}

// TODO Support more characters
var patternRegex = regexp.MustCompile("/\\{([a-z_]+)\\}")

func extractParameters(pattern string) []string {
	// TODO Validate pattern
	parameterElements := patternRegex.FindAllStringSubmatch(pattern, -1)

	parameters := make([]string, len(parameterElements))
	for idx, element := range parameterElements {
		parameters[idx] = element[1]
	}

	return parameters
}

var argumentRegex = regexp.MustCompile("/([A-Za-z_]+)")

func extractPathParameters(argumentPath string, parameters []string) PathParameters {
	// TODO Validate elements against expected parameters OR pass in remaining values in list. Perhaps use "*" to allow this
	argumentElements := argumentRegex.FindAllStringSubmatch(argumentPath, -1)

	pathParams := parameterMap(make(map[string]string, len(parameters)))
	for idx, element := range parameters {
		pathParams[element] = argumentElements[idx][1]
	}

	return pathParams
}

func SingletonResource(i interface{}, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	log.Printf("Registering GET handler for [%s] as [%s]\n", t.String(), name)
	defaultHandlerMutex.registerHandler(name, make([]string, 0), handler)
}

func Resource(i interface{}, parameterPattern string, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	parameters := extractParameters(parameterPattern)
	log.Printf("Registering GET handler for [%s] as [%s] with [%s]\n", t.String(), name, parameterPattern)
	defaultHandlerMutex.registerHandler(name, parameters, handler)
}

func GetResource(r *http.Request) (interface{}, *RequestError) {
	// TODO Handle invalid URLs when determining typeName and suffix. Note, we should always have a leading "/"
	typeName := strings.Trim(strings.SplitAfterN(r.URL.Path, "/", 3)[1], "/")
	argumentPath := strings.TrimPrefix(r.URL.Path, "/" + typeName)
	log.Printf("GET request for [%v] [%v]\n", typeName, argumentPath)

	handler, parameters := defaultHandlerMutex.getHandler(typeName)
	if handler == nil {
		log.Printf("No handler registered for %s", typeName)
		return nil, &RequestError{Error: fmt.Errorf("No handler registered for %s", typeName), Message: "Invalid resource type", Code: http.StatusNotFound}
	}
	log.Printf("Found GET handler for [%v] with [%v]\n", typeName, parameters)

	pathParameters := extractPathParameters(argumentPath, parameters)

	resource, err := handler(pathParameters)
	return resource, err
}
