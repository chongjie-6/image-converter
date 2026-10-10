package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: body}, nil
}

func handler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler(w http.ResponseWriter, r *http.Request){
	wikiTitle := r.URL.Path[len("/view/"):]
	page, _ := loadPage(wikiTitle)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", page.Title, page.Body)
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	log.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}