package server

import (
	"log"
	"net/http"
)

const (
	StatusInternalServerErrorMessage = "An internal server error has occured."
)

type RequestError struct {
	Error   error
	Message string
	Code    int
}

func internalRequestError(e error) *RequestError {
	return &RequestError{Error: e, Message: StatusInternalServerErrorMessage, Code: http.StatusInternalServerError}
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	// TODO Ensure response fmt is valid before proceeding
	res, err := GetResource(r)
	if err != nil {
		log.Printf("Returning [%d] response [%s]", err.Code, err.Message)
		http.Error(w, err.Message, err.Code)
		return
	}

	MarshallResponse(res, w, r)
}

func init() {
	http.HandleFunc("/", MainHandler)
}
