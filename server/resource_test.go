package server_test

import (
	"fmt"
	. "github.com/cleggatt/gowest/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

// GetHandlers for single instance vs collection tests
func getSingleResourceHandler(_ PathParameters) (interface{}, *RequestError) {
	return book{"Neuromancer", "Gibson, William"}, nil
}

func getResourceCollectionHandler(_ PathParameters) (interface{}, *RequestError) {
	return []book{book{"Neuromancer", "Gibson, William"}, book{"Pandora's Star", "Peter F. Hamilton"}}, nil
}

func errorResourceHandler(_ PathParameters) (interface{}, *RequestError) {
	return nil, &RequestError{Error: fmt.Errorf("errorResourceHandler"), Message: "errorResourceHandler", Code: http.StatusBadRequest}
}

func createParameterAccumulator(acc *map[string]string) GetHandler {
	return func (src PathParameters) (interface{}, *RequestError) {
		for k, v := range src.AsMap() {
			(*acc)[k] = v
		}
		return nil, nil
	}
}

var _ = Describe("GET resource handler", func() {
	AfterEach(func() {
		ClearHandlers()
	})
	Describe("registering a type", func() {
		It("should allow use of a instance", func() {
			// Exercise
			SingletonResource(book{}, getSingleResourceHandler)
			// Verify
			req := request("http://localhost:8080/book?fmt=json")

			res, err := GetResource(req)
			Expect(res).To(Equal(book{"Neuromancer", "Gibson, William"}))
			Expect(err).To(BeNil())
		})
		It("should allow use of a pointer", func() {
			// Exercise
			SingletonResource(new(book), getSingleResourceHandler)
			// Verify
			req := request("http://localhost:8080/book?fmt=json")

			res, err := GetResource(req)
			Expect(res).To(Equal(book{"Neuromancer", "Gibson, William"}))
			Expect(err).To(BeNil())
		})
	})
	Describe("GETting a resource", func() {
		Context("when requesting a non-existent resource", func() {
			It("should return a 404 error", func() {
				// Exercise
				SingletonResource(new(book), getSingleResourceHandler)
				// Verify
				req := request("http://localhost:8080/rook?fmt=json")

				res, err := GetResource(req)
				Expect(err.Error).To(Equal(fmt.Errorf("No handler registered for rook")))
				Expect(err.Message).To(Equal("Invalid resource type"))
				Expect(err.Code).To(Equal(404))
				Expect(res).To(BeNil())
			})
		})
		Context("when dealing with returned resources", func() {
			It("should handle single instances", func() {
				// Setup
				SingletonResource(book{}, getSingleResourceHandler)
				// Exercise
				req := request("http://localhost:8080/book?fmt=json")
				res, err := GetResource(req)
				// Verify
				Expect(res).To(Equal(book{"Neuromancer", "Gibson, William"}))
				Expect(err).To(BeNil())
			})
			It("should handle collections", func() {
				// Setup
				SingletonResource(book{}, getResourceCollectionHandler)
				// Exercise
				req := request("http://localhost:8080/book?fmt=json")
				res, err := GetResource(req)
				// Verify
				Expect(res).To(Equal([]book{book{"Neuromancer", "Gibson, William"}, book{"Pandora's Star", "Peter F. Hamilton"}}))
				Expect(err).To(BeNil())
			})
			It("should handle RequestErrors", func() {
				// Exercise
				// TODO Parameterise errorResourceHandler to make test clearer
				SingletonResource(new(book), errorResourceHandler)
				// Verify
				req := request("http://localhost:8080/book?fmt=json")

				res, err := GetResource(req)
				Expect(err.Error).To(Equal(fmt.Errorf("errorResourceHandler")))
				Expect(err.Message).To(Equal("errorResourceHandler"))
				Expect(err.Code).To(Equal(400))
				Expect(res).To(BeNil())
			})
		})
		// TODO Should return based on type
		Context("when requesting with path parameters", func() {
			// TODO Should complain if invalid params are passed
			// TODO Should complain if params are passed in the wrong order
			// TODO Allow leading "/" and a missing last "/" in the pattern and URL
			// TODO Ignore static string sections eg "{last}/ignore/{first}/"
			It("should pass parameters to the handler by name", func() {
				// Setup
				params := make(map[string]string)
				Resource(book{}, "/{author_last}/{author_first}", createParameterAccumulator(&params))
				// Exercise
				req := request("http://localhost:8080/book/hamilton/peter_f")
				GetResource(req)
				// Verify
				Expect(params).To(Equal(map[string]string {"author_last": "hamilton", "author_first": "peter_f"}))
			})
		})
	})
})
