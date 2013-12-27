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

func getBookHandler() interface{} {
	return book{"Neuromancer", "Gibson, William"}
}

func main() {
	Resource(new(book), getBookHandler)

	go http.ListenAndServe(":8080", nil)

	resp, err := http.Get("http://localhost:8080/book?fmt=json")
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
