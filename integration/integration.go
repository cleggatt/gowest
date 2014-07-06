package main

import (
	"bytes"
	. "github.com/cleggatt/gowest/server"
	"io/ioutil"
	"log"
	"net/http"
)

type book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func getBookHandler(params PathParameters) (interface{}, *RequestError) {
	title, _ := params.Get("title")
	return book{title, "Gibson, William"}, nil
}

func main() {
	Resource(book{}, "/{title}", getBookHandler)

	go http.ListenAndServe(":8080", nil)

	resp, err := http.Get("http://localhost:8080/book/Neuromancer?fmt=json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	expected := []byte("{\"title\":\"Neuromancer\",\"author\":\"Gibson, William\"}")
	if !bytes.Equal(body, expected) {
		log.Fatalf("Expected [%v], Actual [%v]", string(expected), string(body))
	}
}
