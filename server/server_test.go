package server_test

import (
	. "github.com/cleggatt/gowest/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http/httptest"
)

func serverBookHandler(_ PathParameters) (interface{}, *RequestError) {
	return book{"Neuromancer", "Gibson, William"}, nil
}

var _ = Describe("Main handler", func() {
	AfterEach(func() {
		ClearHandlers()
	})
	Describe("handling a GET request", func() {
		Context("for a valid request", func() {
			It("should return a response", func() {
				// Set up
				Resource(new(book), serverBookHandler)
				// Exercise
				req := request("http://localhost:8080/book?fmt=json")
				resp := httptest.NewRecorder()
				MainHandler(resp, req)
				// Verify
				Expect(resp.Body.String()).To(Equal("{\"title\":\"Neuromancer\",\"author\":\"Gibson, William\"}"))
			})
		})
		Context("with an error", func() {
			It("should write the correctly formatted response", func() {
				// Exercise
				req := request("http://localhost:8080/Missing?fmt=json")
				resp := httptest.NewRecorder()
				MainHandler(resp, req)
				// Verify
				Expect(resp.Code).To(Equal(404))
				Expect(resp.Body.String()).To(Equal("Invalid resource type\n"))
			})
		})
	})
})
