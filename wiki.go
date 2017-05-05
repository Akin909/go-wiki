package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Page is a struct representing contents of the html page
type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "index.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func homeHandler(write http.ResponseWriter, req *http.Request, tmpl string) {
	// http.Redirect(write, req, "/edit/new", http.StatusFound)
	page := &Page{Title: "Home"}
	myTemplate, err := template.ParseFiles("index.html")
	if err != nil {
		http.Redirect(write, req, "/edit/new", http.StatusFound)
	}
	myTemplate.Execute(write, page)
}

func viewHandler(write http.ResponseWriter, req *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(write, req, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(write, "view", page)

}

func editHandler(write http.ResponseWriter, req *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(write, "edit", page)

}

func saveHandler(write http.ResponseWriter, req *http.Request, title string) {
	body := req.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(write, req, "/view/"+title, http.StatusFound)
}

func renderTemplate(write http.ResponseWriter, tmpl string, page *Page) {
	err := templates.ExecuteTemplate(write, tmpl+".html", page)
	if err != nil {
		http.Error(write, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if r.URL.Path != "/" {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return
			}
			fn(w, r, m[2])
		} else {
			fn(w, r, r.URL.Path)
		}
	}
}

func main() {
	http.HandleFunc("/", makeHandler(homeHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	fmt.Println("Server up on port 8080")
	http.ListenAndServe(":8080", nil)
}
