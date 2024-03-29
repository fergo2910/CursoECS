package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func handleSaveBook(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("web.request", tracer.ResourceName("/save"))
	defer span.Finish()
	
	var id = 0
	var err error

	r.ParseForm()
	params := r.PostForm
	idStr := params.Get("id")

	if len(idStr) > 0 {
		id, err = strconv.Atoi(idStr)
		if err != nil {
			renderErrorPage(w, err)
			return
		}
	}

	name := params.Get("name")
	author := params.Get("author")

	pagesStr := params.Get("pages")
	pages := 0
	if len(pagesStr) > 0 {
		pages, err = strconv.Atoi(pagesStr)
		if err != nil {
			renderErrorPage(w, err)
			return
		}
	}

	publicationDateStr := params.Get("publicationDate")
	var publicationDate time.Time

	if len(publicationDateStr) > 0 {
		publicationDate, err = time.Parse("2006-01-02", publicationDateStr)
		if err != nil {
			renderErrorPage(w, err)
			return
		}
	}

	if id == 0 {
		_, err = insertBook(name, author, pages, publicationDate)
	} else {
		_, err = updateBook(id, name, author, pages, publicationDate)
	}

	if err != nil {
		renderErrorPage(w, err)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func handleListBooks(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("web.request", tracer.ResourceName("/"))
	defer span.Finish()
	
	books, err := allBooks()
	if err != nil {
		renderErrorPage(w, err)
		return
	}

	buf, err := ioutil.ReadFile("www/index.html")
	if err != nil {
		renderErrorPage(w, err)
		return
	}

	var page = IndexPage{AllBooks: books}
	indexPage := string(buf)
	t := template.Must(template.New("indexPage").Parse(indexPage))
	t.Execute(w, page)
}

func handleViewBook(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("web.request", tracer.ResourceName("/book"))
	defer span.Finish()
		
	params := r.URL.Query()
	idStr := params.Get("id")

	var currentBook = Book{}
	currentBook.PublicationDate = time.Now()

	if len(idStr) > 0 {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			renderErrorPage(w, err)
			return
		}

		currentBook, err = getBook(id)
		if err != nil {
			renderErrorPage(w, err)
			return
		}
	}

	buf, err := ioutil.ReadFile("www/book.html")
	if err != nil {
		renderErrorPage(w, err)
		return
	}

	var page = BookPage{TargetBook: currentBook}
	bookPage := string(buf)
	t := template.Must(template.New("bookPage").Parse(bookPage))
	err = t.Execute(w, page)
	if err != nil {
		renderErrorPage(w, err)
		return
	}
}

func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("web.request", tracer.ResourceName("/delete"))
  defer span.Finish()

	params := r.URL.Query()
	idStr := params.Get("id")

	if len(idStr) > 0 {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			renderErrorPage(w, err)
			return
		}

		n, err := removeBook(id)
		if err != nil {
			renderErrorPage(w, err)
			return
		}

		fmt.Printf("Rows removed: %v\n", n)
	}
	http.Redirect(w, r, "/", 302)
}

func renderErrorPage(w http.ResponseWriter, errorMsg error) {
	buf, err := ioutil.ReadFile("www/error.html")
	if err != nil {
		log.Printf("%v\n", err)
		fmt.Fprintf(w, "%v\n", err)
		return
	}

	var page = ErrorPage{ErrorMsg: errorMsg.Error()}
	errorPage := string(buf)
	t := template.Must(template.New("errorPage").Parse(errorPage))
	t.Execute(w, page)
}

func check(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan("web.request", tracer.ResourceName("/check"))
  defer span.Finish()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}