package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	hTemplate "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	tTemplate "text/template"
)

type template interface {
	Execute(wr io.Writer, data interface{}) (err error)
}

type anyTemplate struct {
	template
}

func fmtType(i interface{}) string {
	v := reflect.ValueOf(i)
	return v.Type().String()
}

func (at anyTemplate) execute(i interface{}, w io.Writer) error {
	return at.template.Execute(w, i)
}

func parseHtmlTemplate(filename string) (anyTemplate, error) {
	var t, err = hTemplate.ParseFiles(filename)
	return anyTemplate{t}, err
}

func parseTextTemplate(filename string) (anyTemplate, error) {
	var t, err = tTemplate.ParseFiles(filename)
	return anyTemplate{t}, err
}

func loadTemplate(i interface{}, format string) (anyTemplate, error) {
	typeName := fmtType(i)
	if format == "html" {
		return parseHtmlTemplate(typeName + ".html")
	} else {
		// Since we don't know anything about these formats, we need to rely on the template to do what's right
		// e.g. for a CSV the template should enclose all fields in double quotes to handle special characters
		// FIXME Security risk - using client data input
		// TODO Add in support for a list of supported formats, the use of which should be recommended
		// TODO If taking the format from the client (i.e. there is no list of valid formats), lower case it
		return parseTextTemplate(typeName + "." + format)
	}
}

func createLoadTemplateError(format string, err error) *RequestError {
	// FIXME We're assuming that any PathError will be caused by the file not existing - we should in fact
	// explicitly check for the existence of the file
	if perr, ok := err.(*os.PathError); ok {
		log.Printf("Template does not exist: %v", perr)
		// FIXME Security risk - using client data input
		// TODO List supported formats in response
		return &RequestError{Error: perr, Message: fmt.Sprintf("'%s' is not a supported format", format), Code: http.StatusNotAcceptable}
	} else {
		// TODO When we fix the above, add a test for this case
		log.Printf("Unable to load template: %v", err)
		return internalRequestError(err)
	}
}

func MarshallResponse(i interface{}, wr io.Writer, r *http.Request) *RequestError {
	// If the fmt parameter appears twice, we take the first one
	// TODO Default to first format in list
	format := r.URL.Query().Get("fmt")
	if format == "json" {
		bytes, err := json.Marshal(i)
		if err != nil {
			log.Printf("Unable to marshall instance of [%v] ([%v]): %v", fmtType(i), i, err)
			return internalRequestError(err)
		}
		if _, err := wr.Write(bytes); err != nil {
			// At this point, it's likely we won't be able to write this internal service error anyway
			log.Printf("Unable to write response [%v]: %v", string(bytes), err)
			return internalRequestError(err)
		}
	} else {
		template, err := loadTemplate(i, format)
		if err != nil {
			return createLoadTemplateError(format, err)
		}
		// TODO Allow the disabling of this format checking for efficiency (if desired)
		// TODO Allocate the buffer to be the same size of the template
		// The execution process writes directly to the buffer, so it may write bytes before finding an error
		buff := new(bytes.Buffer)
		if err := template.execute(i, buff); err != nil {
			log.Printf("Unable to process template [%v]", err)
			return internalRequestError(err)
		}
		if _, err := wr.Write(buff.Bytes()); err != nil {
			// At this point, it's likely we won't be able to write this internal service error anyway
			log.Printf("Unable to write response [%v]: %v", buff.String(), err)
			return internalRequestError(err)
		}
	}
	return nil
}
