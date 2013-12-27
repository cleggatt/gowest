package server_test

import (
	. "github.com/cleggatt/gowest/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"net/http"
	"os"
)

type badJson struct {
	InvalidType func()
}

type badWriter struct {
}

func (b *badWriter) Write(p []byte) (n int, err error) {
	return -1, fmt.Errorf("Bad writer!")
}

var _ = Describe("Representation", func() {
	Describe("Marshalling a resource", func() {
		cases := map[string]string{
			"json": "{\"title\":\"Neuromancer\",\"author\":\"Gibson, William\"}",
			"html": "<html><body>Neuromancer by Gibson, William</body></html>",
			"text": "Neuromancer by Gibson, William",
			"csv":  "\"Neuromancer\",\"Gibson, William\""}
		for k, v := range cases {
			format, expected := k, v
			Context("as "+format, func() {
				It("should write the correctly formatted response", func() {
					// Exercise
					req := request("http://localhost:8080/book?fmt=" + format)
					resp := new(bytes.Buffer)
					err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
					// Verify
					Expect(resp.String()).To(Equal(expected))
					Expect(err).To(BeNil())
				})
			})
		}
		Context("to an unsupported format", func() {
			It("should return a 406 error", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=missing")
				resp := new(bytes.Buffer)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err.Code).To(Equal(http.StatusNotAcceptable))
				Expect(err.Message).To(Equal("'missing' is not a supported format"))
				Expect(resp.String()).To(Equal(""))
				Expect(err.Error).To(BeAssignableToTypeOf(new(os.PathError)))
			})
		})
		Context("with struct that's invalid for JSON marshalling", func() {
			It("should return a 500 error", func() {
				// Exercise
				req := request("http://localhost:8080/badJson?fmt=json")
				resp := new(bytes.Buffer)
				err := MarshallResponse(badJson{}, resp, req)
				// Verify
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
		})
		Context("with an bad template", func() {
			It("should return a 500 error", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=bad")
				resp := new(bytes.Buffer)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
		})
		Context("with a error when writing a response", func() {
			It("should return a 500 error for JSON", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=json")
				resp := new(badWriter)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
			})
			It("should return a 500 error for a template", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=html")
				resp := new(badWriter)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
			})
		})
	})
})
