package server_test

import (
	. "github.com/cleggatt/gowest/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"net/http"
)

type badJson struct {
	InvalidType func()
}

type noHtmlTemplate struct {
}

type errorExecute struct {
	Foo string;
}

type errorParse struct {
	Foo string;
}

type badWriter struct {
}

func (b *badWriter) Write(p []byte) (n int, err error) {
	return -1, fmt.Errorf("Bad writer!")
}

var _ = Describe("representation.go", func() {
	Describe("Generating a representation", func() {
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
					Expect(resp).ToNot(BeNil())
					Expect(resp.String()).To(Equal(expected))
					Expect(err).To(BeNil())
				})
			})
		}
		Context("when there is no corresponding template", func() {
			It("should return a 406 error for HTML formats", func() {
				// Exercise
				req := request("http://localhost:8080/noHtmlTemplate?fmt=html")
				resp := new(bytes.Buffer)
				err := MarshallResponse(noHtmlTemplate{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusNotAcceptable))
				Expect(err.Message).To(Equal("'html' is not a supported format"))
				Expect(resp.String()).To(Equal(""))
			})
			It("should return a 406 error for non-HTML formats", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=missing")
				resp := new(bytes.Buffer)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusNotAcceptable))
				Expect(err.Message).To(Equal("'missing' is not a supported format"))
				Expect(resp.String()).To(Equal(""))
			})
		})
		Context("when resource rendering fails", func() {
			It("should return a 500 error when due to JSON marshalling failure", func() {
				// Exercise
				req := request("http://localhost:8080/badJson?fmt=json")
				resp := new(bytes.Buffer)
				err := MarshallResponse(badJson{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
			It("should return a 500 error when due to HTML template parsing failure", func() {
				// Exercise
				req := request("http://localhost:8080/errorParse?fmt=text")
				resp := new(bytes.Buffer)
				err := MarshallResponse(errorParse{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
			It("should return a 500 error when due to non-HTML template parsing failure", func() {
				// Exercise
				req := request("http://localhost:8080/errorParse?fmt=text")
				resp := new(bytes.Buffer)
				err := MarshallResponse(errorParse{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
			It("should return a 500 error when due to HTML template execution failure", func() {
				// Exercise
				req := request("http://localhost:8080/errorExecute?fmt=html")
				resp := new(bytes.Buffer)
				err := MarshallResponse(errorExecute{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
			It("should return a 500 error when due to non-HTML template execution failure", func() {
				// Exercise
				req := request("http://localhost:8080/errorExecute?fmt=text")
				resp := new(bytes.Buffer)
				err := MarshallResponse(errorExecute{}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
				Expect(resp.String()).To(Equal(""))
			})
		})
		Context("when unable to writing a response", func() {
			It("should return a 500 error for JSON responses", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=json")
				resp := new(badWriter)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
			})
			It("should return a 500 error for a HTML template response", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=html")
				resp := new(badWriter)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
			})
			It("should return a 500 error for a non-HTML template response", func() {
				// Exercise
				req := request("http://localhost:8080/book?fmt=text")
				resp := new(badWriter)
				err := MarshallResponse(book{"Neuromancer", "Gibson, William"}, resp, req)
				// Verify
				Expect(err).ToNot(BeNil())
				Expect(err.Code).To(Equal(http.StatusInternalServerError))
				Expect(err.Message).To(Equal(StatusInternalServerErrorMessage))
			})
		})
	})
})
