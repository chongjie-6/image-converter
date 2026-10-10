package main

import (
	// "fmt"
	"log"
	// "net/http"
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

func main(){
	testWikiPage := &Page{Title: "TestPage", Body: []byte("This is the body of the test page!")}
	error := testWikiPage.save()

	if error != nil {
		log.Fatal("Error saving test page: ", error)
		return
	}

	log.Printf("Test page saved successfully with title '%s'", testWikiPage.Title)

	page, error := loadPage("TestPage")
	
	if error != nil {
		log.Fatal("Error loading test page: ", error)
		return
	}

	log.Printf("Test page loaded successfully with title '%s' and body: '%s'", page.Title, page.Body)
}

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, World!")
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	log.Println("Server is running on http://localhost:3000")
// 	log.Fatal(http.ListenAndServe(":3000", nil))
// }