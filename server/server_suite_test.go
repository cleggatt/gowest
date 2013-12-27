package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func request(rawurl string) *http.Request {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		panic(fmt.Sprintf("Unable to create URL for [%s]", rawurl))
	}
	return &http.Request{URL: url}
}

type book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Server Suite")
}
