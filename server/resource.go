package server

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

type GetHandler func() interface{}

// TODO Fix naming i.e. handlerMutex.mutex
type handlerMutex struct {
	mutex    sync.RWMutex
	handlers map[string]mutexEntry
}

type mutexEntry struct {
	handler  GetHandler
	typeName string
}

func newHandlerMutex() *handlerMutex { return &handlerMutex{handlers: make(map[string]mutexEntry)} }

func (mutex *handlerMutex) registerHandler(typeName string, handler GetHandler) {
	mutex.mutex.Lock()
	defer mutex.mutex.Unlock()

	mutex.handlers[typeName] = mutexEntry{handler: handler, typeName: typeName}
}

func (mutex *handlerMutex) getHandler(typeName string) GetHandler {
	mutex.mutex.RLock()
	defer mutex.mutex.RUnlock()

	// TODO Handle missing map entry
	return mutex.handlers[typeName].handler
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

func Resource(i interface{}, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	log.Printf("Registering GET handler for [%s] as [%s]\n", t.String(), name)
	defaultHandlerMutex.registerHandler(name, handler)
}

func GetResource(r *http.Request) (interface{}, *RequestError) {
	// TODO Handle invalid URL
	typeName := strings.TrimPrefix(r.URL.Path, "/")
	log.Printf("GET request for [%v]\n", typeName)

	handler := defaultHandlerMutex.getHandler(typeName)
	if handler == nil {
		log.Printf("No handler registered for %s", typeName)
		return nil, &RequestError{Error: fmt.Errorf("No handler registered for %s", typeName), Message: "Invalid resource type", Code: http.StatusNotFound}
	}

	log.Printf("Found GET handler for [%v]\n", typeName)
	return handler(), nil
}
