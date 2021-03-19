package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]*)$")
var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html", "templates/index.html"))
var body_DIR string = "body_text/"
var template_DIR string = "templates/"

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	fmt.Println(p.Title)
	renderTemplate(w, "view", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles(template_DIR + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(body_DIR+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := body_DIR + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fmt.Println("makehandler:", m)
		fn(w, r, m[2])
	}
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if filepath.Ext(path) == ".txt" {
			*files = append(*files, path)
		}
		return nil
	}
}
func loadPages(file_path string, pages *[]Page) (*Page, error) {
	filename := file_path
	body, err := ioutil.ReadFile(filename)
	full_path := strings.Split(file_path, "/")
	var f_name string = full_path[1]
	var title string = f_name[0 : len(f_name)-4]
	if err != nil {
		return nil, err
	}
	*pages = append(*pages, Page{Title: title, Body: body})
	return &Page{Title: title, Body: body}, nil
}

func mainpageHandler(w http.ResponseWriter, r *http.Request) {
	var files []string
	var pages []Page
	web_root := body_DIR
	err := filepath.Walk(web_root, visit(&files))
	if err == nil {
		for i := 0; i < len(files); i++ {
			fmt.Println(files[i])
			p, errr := loadPages(files[i], &pages)
			if errr != nil {
				fmt.Println(p)
				return
			}
		}
		err := templates.ExecuteTemplate(w, "index.html", pages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func main() {
	argsWithoutProg := os.Args[1:]
	fmt.Print(argsWithoutProg)
	Os_DIR_static := "./static"
	Static_Prefix := "/static/"
	fs := http.StripPrefix(Static_Prefix, http.FileServer(http.Dir(Os_DIR_static)))
	http.Handle(Static_Prefix, fs)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", mainpageHandler)
	log.Fatal(http.ListenAndServe(":8008", nil))
}
