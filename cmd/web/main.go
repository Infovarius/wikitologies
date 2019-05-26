package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/stillpiercer/wikitologies/graph"
	"github.com/stillpiercer/wikitologies/parser"
	wikt "github.com/stillpiercer/wikitologies/wiktionary"
)

const (
	PNG = "png"
	SVG = "svg"
)

var (
	mainTemplate *template.Template
	viewTemplate *template.Template
	editTemplate *template.Template
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mainTemplate = template.Must(template.ParseFiles(wd + "/templates/main.html"))
	viewTemplate = template.Must(template.
		New("view.html").
		Funcs(template.FuncMap{"svg": svg}).
		ParseFiles(wd + "/templates/view.html"))
	editTemplate = template.Must(template.ParseFiles(wd + "/templates/edit.html"))

	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/{title}", viewHandler)
	r.HandleFunc("/edit/{title}", editHandler)
	r.HandleFunc("/save/{title}", saveHandler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), recovery(r)); err != nil {
		panic(err)
	}
}

func mainHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := mainTemplate.Execute(w, nil); err != nil {
		panic(err)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(mux.Vars(r)["title"], ":")
	title, lang := split[0], ""
	if len(split) > 1 {
		lang = split[1]
	}

	params := make(map[string]int)
	for k, v := range r.URL.Query() {
		value, err := strconv.Atoi(v[0])
		if err != nil {
			continue
		}

		params[k] = value
	}

	var strict bool
	if r.URL.Query().Get("strict") == "true" {
		strict = true
	}

	data := struct {
		Title  string
		Lang   string
		Strict bool
		Params map[string]int
	}{
		Title:  title,
		Lang:   lang,
		Strict: strict,
		Params: params,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := viewTemplate.Execute(w, data); err != nil {
		panic(err)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(mux.Vars(r)["title"], ":")
	title, lang := split[0], ""
	if len(split) > 1 {
		lang = split[1]
	}

	text, err := wikt.GetText(title)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	sections := parser.ParseText(text)
	word := parser.NewWord(title, sections)
	var meanings parser.Meanings
	if lang != "" {
		meanings = word.ByLanguage(lang)
	} else {
		meanings = word[0].Meanings
	}
	if meanings == nil {
		_, _ = fmt.Fprintf(w, "language %s for %s not found", lang, title)
		return
	}

	var values []string
	for _, m := range meanings {
		values = append(values, m.Value)
	}

	data := struct {
		Title  string
		Values []string
	}{
		Title:  title,
		Values: values,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := editTemplate.Execute(w, data); err != nil {
		panic(err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = PNG
	}

	w.Header().Set("Content-Disposition", "attachment; filename=WHATEVER_YOU_WANT")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
}

func svg(title, lang string, strict bool, params map[string]int) template.HTML {
	g, err := graph.Build(title, lang, strict, params)
	if err != nil {
		return template.HTML(err.Error())
	}

	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = strings.NewReader(g.String())

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return template.HTML(out)
}

func recovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprint(r), http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
