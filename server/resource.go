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
	pattern string
	handler GetHandler
}

func newHandlerMutex() *handlerMutex { return &handlerMutex{handlers: make(map[string]mutexEntry)} }

func (mutex *handlerMutex) registerHandler(typeName string, pattern string, handler GetHandler) {
	mutex.mutex.Lock()
	defer mutex.mutex.Unlock()

	mutex.handlers[typeName] = mutexEntry{typeName: typeName, pattern: pattern, handler: handler }
}

// FIXME Does naming the string actually make the signature clearer? Confirm that it appears in godocs
// FIXME e.g. It would be nice to describe this has (GetHandler, pattern string)
// FIXME BUT Using named seems to crate temp variables unncessarily see  getInterfaceTypeName()
func (mutex *handlerMutex) getHandler(typeName string) (GetHandler, string) {
	mutex.mutex.RLock()
	defer mutex.mutex.RUnlock()

	// TODO Handle missing map entry
	entry := mutex.handlers[typeName]
	return entry.handler, entry.pattern
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

var patternRegex = regexp.MustCompile("\\{([a-z_]+)\\}/")
var urlRegex = regexp.MustCompile("([a-z_]+)/")

func extractParameters(parameterPath string, pattern string) PathParameters {
	// TODO Validate url elements against expected OR pass in remaining values in list. Perhaps use "*" to allow this
	patternElements := patternRegex.FindAllStringSubmatch(pattern, -1)
	parameterElements := urlRegex.FindAllStringSubmatch(parameterPath, -1)

	pathParams := parameterMap(make(map[string]string, len(patternElements)))
	for idx, element := range patternElements {
		pathParams[element[1]] = parameterElements[idx][1]
	}

	return pathParams
}

func Resource(i interface{}, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	log.Printf("Registering GET handler for [%s] as [%s]\n", t.String(), name)
	defaultHandlerMutex.registerHandler(name, "", handler)
}

func ParameterisedResource(i interface{}, pattern string, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	log.Printf("Registering GET handler for [%s] as [%s] with [%s]\n", t.String(), name, pattern)
	defaultHandlerMutex.registerHandler(name, pattern, handler)
}

func GetResource(r *http.Request) (interface{}, *RequestError) {
	// TODO Handle invalid URLs when determining typeName and suffix. Note, we should always have a leading "/"
	typeName := strings.Trim(strings.SplitAfterN(r.URL.Path, "/", 3)[1], "/")
	parameterPath := strings.TrimPrefix(r.URL.Path, "/" + typeName + "/")
	log.Printf("GET request for [%v] [%v]\n", typeName, parameterPath)

	handler, pattern := defaultHandlerMutex.getHandler(typeName)
	if handler == nil {
		log.Printf("No handler registered for %s", typeName)
		return nil, &RequestError{Error: fmt.Errorf("No handler registered for %s", typeName), Message: "Invalid resource type", Code: http.StatusNotFound}
	}
	log.Printf("Found GET handler for [%v] with [%v\n", typeName, pattern)

	pathParams := extractParameters(parameterPath, pattern)
	resource, err := handler(pathParams)

	return resource, err
}
