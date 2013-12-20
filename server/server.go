package server

import (
	"encoding/json"
	hTemplate "html/template"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	tTemplate "text/template"
)

func writeResponse(i interface{}, w http.ResponseWriter, r *http.Request) {

	// If the fmt parameter appears twice, we take the first one
	// TODO Default to first format in list
	fmt := r.URL.Query().Get("fmt")

	if fmt == "json" {
		bytes, err := json.Marshal(i)
		if err != nil {
			// TODO Write error response
			return
		}
		// TODO Handle error
		w.Write(bytes)
	} else {

		v := reflect.ValueOf(i)
		fName := v.Type().String()

		// We treat HTML specially, due to the ability to inject malicious code
		if fmt == "html" {
			// TODO Handle error
			t, _ := hTemplate.ParseFiles(fName + ".html")
			t.Execute(w, i)
		} else {
			// Since we don't know anything about these formats, we need to rely on the template to do what's right
			// e.g. for a CSV the template should enclose all fields in double quotes to handle special characters
			// FIXME Security risk - using client data input
			// TODO Add in support for a list of supported formats, the use of which should be recommended
			// TODO If taking the format from the client (i.e. there is no list of valid formats), lower case it
			// TODO handle error
			t, _ := tTemplate.ParseFiles(fName + "." + fmt)
			t.Execute(w, i)
		}
		// TODO Error - unknown format error
		// If the format is on the list but there is no template - server side error (500)
		// If the format is not on the list (or there is no list) - client side error (406)
	}
}

// TODO Add id parameter
type GetHandler func() interface{}

// TODO Fix naming i.e. handlerMutex.mutex
type handlerMutex struct {
	mutex    	sync.RWMutex
	handlers	map[string]mutexEntry
}

type mutexEntry struct {
	handler   GetHandler
	typeName  string
}

func newHandlerMutex() *handlerMutex { return &handlerMutex{handlers: make(map[string]mutexEntry)} }

func (mutex *handlerMutex) registerHandler(typeName string, handler GetHandler) {
	mutex.mutex.Lock()
	defer mutex.mutex.Unlock()

	mutex.handlers[typeName] = mutexEntry{handler: handler, typeName: typeName}
}

func (mutex *handlerMutex) getHandler(typeName string) GetHandler {
	mutex.mutex.Lock()
	defer mutex.mutex.Unlock()

	// TODO Handle missing map entry
	return mutex.handlers[typeName].handler
}

func clearHandlers() {
	// TODO This is not threadsafe, but is just used for tests ATM
	defaultHandlerMutex = newHandlerMutex()
}

var defaultHandlerMutex = newHandlerMutex()

func getInterfaceTypeName(i interface{}) (t reflect.Type, name string) {
	t = reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr  {
		pointerToValue := reflect.ValueOf(i)
		valueAtPointer := reflect.Indirect(pointerToValue);
		t = valueAtPointer.Type();
	}
	name = t.Name()
	return
}

func Resource(i interface{}, handler GetHandler) {
	t, name := getInterfaceTypeName(i)
	log.Printf("Registering GET handler for [%v] as [%v]\n", t, name)
	defaultHandlerMutex.registerHandler(name, handler)
}

func getResource(r *http.Request) interface{} {
	// TODO Handle invalid URL
	typeName := strings.TrimPrefix(r.URL.Path,"/")
	log.Printf("GET request for [%v]\n", typeName)
	// TODO Handle missing handler
	handler := defaultHandlerMutex.getHandler(typeName)
	log.Printf("Found GET handler for [%v]\n", typeName)
	return handler()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	res := getResource(r);
	writeResponse(res, w, r)
}

func init() {
	http.HandleFunc("/", mainHandler)
}
