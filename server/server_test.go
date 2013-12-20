package server

import (
	. "github.com/franela/goblin"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func request(rawurl string) *http.Request {
	// TODO Handle error
	url, _ := url.ParseRequestURI(rawurl)
	return &http.Request{URL: url}
}

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func bookHandler() interface{} {
	return Book{"Neuromancer", "Gibson, William"}
}

func TestResourceRegistration(t *testing.T) {
	g := Goblin(t)
	g.Describe("Registering a resource handler", func() {
			g.AfterEach(func() { clearHandlers() })
			g.It("Should use the correct handler when registering using a value", func() {
					// Exercise
					Resource(Book{}, bookHandler)
					// Verify
					req := request("http://localhost:8080/Book?fmt=json")
					res := getResource(req)

					if book, ok := res.(Book); ok {
						g.Assert(book.Author).Equal("Gibson, William")
						g.Assert(book.Title).Equal("Neuromancer")
					} else {
						g.Fail("Result is not a Book")
					}
			})
			g.It("Should use the correct handler when registering using a pointer", func() {
					// Exercise
					Resource(new(Book), bookHandler)
					// Verify
					req := request("http://localhost:8080/Book?fmt=json")
					res := getResource(req)

					if book, ok := res.(Book); ok {
						g.Assert(book.Author).Equal("Gibson, William")
						g.Assert(book.Title).Equal("Neuromancer")
					} else {
						g.Fail("Result is not a Book")
					}
				})
	})
}

func TestResponseFormatting(t *testing.T) {
	g := Goblin(t)
	g.Describe("Writing a response", func() {

			cases := map[string]string{
			"json": "{\"title\":\"Neuromancer\",\"author\":\"Gibson, William\"}",
			"html": "<html><body>Neuromancer by Gibson, William</body></html>",
			"text": "Neuromancer by Gibson, William",
			"csv":  "\"Neuromancer\",\"Gibson, William\""}

			for k, v := range cases {
				format, expected := k, v
				g.It("Should support "+format+" when the corresponding template exists", func() {
						// Exercise
						req := request("http://localhost:8080/Book?fmt=" + format)
						resp := httptest.NewRecorder()
						writeResponse(Book{"Neuromancer", "Gibson, William"}, resp, req)
						// Verify
						g.Assert(resp.Body.String()).Equal(expected)
					})
			}
		})
}

func TestIntegration(t *testing.T) {
	g := Goblin(t)
	g.Describe("Integration testing", func() {
			g.AfterEach(func() { clearHandlers() })
			g.It("Should handle a successful end to end GET request", func() {
					// Set up
					Resource(new(Book), bookHandler)
					// Exercise
					req := request("http://localhost:8080/Book?fmt=json")
					resp := httptest.NewRecorder()
					mainHandler(resp, req)
					// Verify
					g.Assert(resp.Body.String()).Equal("{\"title\":\"Neuromancer\",\"author\":\"Gibson, William\"}")
				})
		})
}
