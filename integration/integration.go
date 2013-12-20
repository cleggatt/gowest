package main

import (
	"bytes"
	"io/ioutil"
	. "github.com/cleggatt/gowest/server"
	"log"
	"net/http"
)

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func getBookHandler() interface{} {
	return Book{"Neuromancer", "Gibson, William"}
}

func main() {
	Resource(new(Book), getBookHandler);

	go http.ListenAndServe(":8080", nil)

	resp, err := http.Get("http://localhost:8080/Book?fmt=json")
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
