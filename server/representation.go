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

func parseHtmlTemplate(filename string) (*anyTemplate, error) {
	var t, err = hTemplate.ParseFiles(filename)
	return &anyTemplate{t}, err
}

func parseTextTemplate(filename string) (*anyTemplate, error) {
	var t, err = tTemplate.ParseFiles(filename)
	return &anyTemplate{t}, err
}

func parseTemplate(i interface{}, format string, filename string) (*anyTemplate, error) {
	// FIXME check extension for all html variants htm HTML xhtml OR use config list?
	if format == "html" {
		return parseHtmlTemplate(filename)
	} else {
		// Since we don't know anything about these formats, we need to rely on the template to do what's right
		// e.g. for a CSV the template should enclose all fields in double quotes to handle special characters
		return parseTextTemplate(filename)
	}
}

func loadTemplate(i interface{}, format string) (*anyTemplate, *RequestError) {
	// FIXME Security risk - using client data input
	// TODO Add in support for a list of supported formats, the use of which should be recommended
	// TODO If taking the format from the client (i.e. there is no list of valid formats), lower case it
	filename := fmtType(i) + "." + format;
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("Template [%s] does not exist: %v", filename, err)
		return nil, &RequestError{Error: err, Message: fmt.Sprintf("'%s' is not a supported format", format), Code: http.StatusNotAcceptable}
	}
	t, err := parseTemplate(i, format, filename);
	if err != nil {
		log.Printf("Unable to parse template: %v", err)
		return nil, internalRequestError(err)
	}
	return t, nil
}

func getJsonBytes(i interface{}) ([]byte, *RequestError) {
	bytes, err := json.Marshal(i)
	if err != nil {
		log.Printf("Unable to marshall instance of [%v] ([%v]): %v", fmtType(i), i, err)
		return nil, internalRequestError(err)
	}
	return bytes, nil
}

func getTemplateBytes(i interface{}, format string) ([]byte, *RequestError) {
	template, err := loadTemplate(i, format)
	if err != nil {
		return nil, err;
	}
	// TODO Allocate the buffer to be the same size of the template
	// The execution process writes directly to the buffer, so it may write bytes before finding an error
	buff := new(bytes.Buffer)
	if err := template.execute(i, buff); err != nil {
		log.Printf("Unable to process template [%v]", err)
		return nil, internalRequestError(err)
	}
	return buff.Bytes(), nil
}

func getBytes(i interface{}, r *http.Request) ([]byte, *RequestError) {
	// TODO Default to first format in list if none is specified
	// If the fmt parameter appears twice, we take the first one
	if format := r.URL.Query().Get("fmt"); format != "json" {
		return getTemplateBytes(i, format)
	} else {
		return getJsonBytes(i)
	}
}

func MarshallResponse(i interface{}, wr io.Writer, r *http.Request) *RequestError {
	bytes, err := getBytes(i, r)
	if err != nil {
		return err
	}
	if _, err := wr.Write(bytes); err != nil {
		// At this point, it's likely we won't be able to write this internal service error anyway
		log.Printf("Unable to write response [%v]: %v", string(bytes), err)
		return internalRequestError(err)
	}
	return nil
}
