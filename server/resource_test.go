package server_test

import (
	"fmt"
	. "github.com/cleggatt/gowest/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func resourceBookHandler() interface{} {
	return book{"Neuromancer", "Gibson, William"}
}

var _ = Describe("GET resource handler", func() {
	AfterEach(func() {
		ClearHandlers()
	})
	Describe("registration", func() {
		Context("when using a instance", func() {
			It("should use a path matching the resource type", func() {
				// Exercise
				Resource(book{}, resourceBookHandler)
				// Verify
				req := request("http://localhost:8080/book?fmt=json")

				res, err := GetResource(req)
				Expect(res).To(Equal(book{"Neuromancer", "Gibson, William"}))
				Expect(err).To(BeNil())
			})
		})
		Context("when using a pointer to an instance", func() {
			It("should use a path matching the resource type", func() {
				// Exercise
				Resource(new(book), resourceBookHandler)
				// Verify
				req := request("http://localhost:8080/book?fmt=json")

				res, err := GetResource(req)
				Expect(res).To(Equal(book{"Neuromancer", "Gibson, William"}))
				Expect(err).To(BeNil())
			})
		})
	})
	Describe("error handling", func() {
		Context("for a path with no handler", func() {
			It("should return a 404 error", func() {
				// Exercise
				Resource(new(book), resourceBookHandler)
				// Verify
				req := request("http://localhost:8080/rook?fmt=json")

				res, err := GetResource(req)
				Expect(err.Error).To(Equal(fmt.Errorf("No handler registered for rook")))
				Expect(err.Message).To(Equal("Invalid resource type"))
				Expect(err.Code).To(Equal(404))
				Expect(res).To(BeNil())
			})
		})
	})
})
