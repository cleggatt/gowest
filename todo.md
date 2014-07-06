Basic Handler examples
----------------------

/books
/books/{isbn}
/books?title={query}

/author
/author/{surname}
/author/{surname}/{firstname}
/author/{surname}/{firstname}/{index}

/books -> BookHandler().get()
/books?title={query} -> BookHandler().get(title->{query})
/books/{isbn} -> BookHandler().get({isbn})

/author -> AuthorHandler().get()
/author/{surname} -> AuthorHandler().get({surname})
/author/{surname}/{firstname} -> AuthorHandler().get({surname}, {firstname})
/author/{surname}/{firstname}/{index} -> AuthorHandler().get({surname}, {firstname}, {index})

Nested resources parts
----------------------

/books/{isbn}/publisher

/books/{isbn} -> BookHandler().get({isbn}).publisher()

Nested Handlers
---------------

/author/{surname}/{firstname}/{index}/books

/author/{surname}/{firstname}/{index}/books -> AuthorHandler().get({surname}, {firstname}, {index}).books()

  val a = AuthorHandler().get({surname}, {firstname}, {index})
  var b = BookHandler().get(a)

Note: If AuthorHandler().get({surname}, {firstname}, {index}).books() exists, it will be used in preference


Note for non-GET requests
-------------------------

If we allowed the following resource

 /author/{id}/books/{isbn}

(which would actually not exists? Maybe /author/{id}/books/ should return hypermedia links?)

then

DELETE /author/{id}/books/{isbn}

This would be

  val a = AuthorHandler().get({id})
  var b = BookHandler().delete(isbn)