package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"

	"github.com/stillpiercer/wikitologies/graph"
	"github.com/stillpiercer/wikitologies/parser"
)

const (
	SVG = "svg"
	DOT = "dot"

	defaultPort  = "8080"
	defaultRedis = "6379"
)

var (
	mainTemplate *template.Template
	viewTemplate *template.Template
	editTemplate *template.Template

	pool *redis.Pool
)

func main() {
	initRedis()
	defer pool.Close()

	wd, err := os.Getwd()
	panicIf(err)

	mainTemplate = template.Must(template.ParseFiles(wd + "/templates/main.html"))
	viewTemplate = template.Must(template.
		New("view.html").
		Funcs(template.FuncMap{"draw": draw}).
		ParseFiles(wd + "/templates/view.html"))
	editTemplate = template.Must(template.ParseFiles(wd + "/templates/edit.html"))

	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/{titles}", viewHandler)
	r.HandleFunc("/edit/{title}", editHandler)
	r.HandleFunc("/save/{format}/{titles}", saveHandler)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}
	log.Println("listening on", port)
	err = http.ListenAndServe(":"+port, recovery(r))
	panicIf(err)
}

func initRedis() {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			var c redis.Conn
			var err error
			url, ok := os.LookupEnv("REDIS_URL")
			if ok {
				c, err = redis.DialURL(url)
			} else {
				c, err = redis.Dial("tcp", ":"+defaultRedis)
			}
			panicIf(err)
			return c, nil
		},
	}
}

func mainHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := mainTemplate.Execute(w, parser.Languages.Names)
	panicIf(err)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	titles, lang := parseTitlesLang(r)
	strict, params := parseStrictParams(r)

	data := struct {
		Titles []string
		Lang   string
		Strict bool
		Params map[string]int
	}{
		Titles: titles,
		Lang:   lang,
		Strict: strict,
		Params: params,
	}

	w.Header().Set("Content-Type", "text/html")
	err := viewTemplate.Execute(w, data)
	panicIf(err)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	split := strings.Split(mux.Vars(r)["title"], "@")
	title, lang := split[0], split[1]

	if strings.Contains(title, "->") {
		title = strings.Split(title, "->")[1]
	}

	word, err := graph.GetWord(title, pool)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	meanings := word.ByLanguage(lang)
	if meanings == nil {
		_, _ = fmt.Fprintf(w, "language %s for %s not found", lang, title)
		return
	}

	data := struct {
		Title    string
		Lang     string
		Meanings parser.Meanings
	}{
		Title:    title,
		Lang:     lang,
		Meanings: meanings,
	}

	w.Header().Set("Content-Type", "text/html")
	err = editTemplate.Execute(w, data)
	panicIf(err)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	titles, lang := parseTitlesLang(r)
	strict, params := parseStrictParams(r)
	format := mux.Vars(r)["format"]

	data, err := dot(titles, lang, strict, params, format)
	panicIf(err)

	filename := fmt.Sprintf("attachment; filename=%s.%s", strings.Join(titles, "+"), format)
	w.Header().Set("Content-Disposition", filename)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	_, err = io.Copy(w, bytes.NewReader(data))
	panicIf(err)
}

func parseTitlesLang(r *http.Request) ([]string, string) {
	split := strings.Split(mux.Vars(r)["titles"], "@")
	return strings.Split(split[0], "+"), split[1]
}

func parseStrictParams(r *http.Request) (bool, map[string]int) {
	var strict bool
	if r.URL.Query().Get("strict") == "true" {
		strict = true
	}

	params := make(map[string]int)
	for k, v := range r.URL.Query() {
		last := len(v) - 1
		value, err := strconv.Atoi(v[last])
		if err != nil {
			continue
		}
		params[k] = value
	}

	return strict, params
}

func dot(titles []string, lang string, strict bool, params map[string]int, format string) ([]byte, error) {
	g, err := graph.Build(titles, lang, strict, params, pool)
	if err != nil {
		return nil, err
	}

	if format == DOT {
		return []byte(g.String()), nil
	}

	cmd := exec.Command("dot", "-T"+format)
	cmd.Stdin = strings.NewReader(g.String())

	return cmd.Output()
}

func draw(titles []string, lang string, strict bool, params map[string]int) template.HTML {
	data, err := dot(titles, lang, strict, params, SVG)
	if err != nil {
		return template.HTML(err.Error())
	}

	return template.HTML(data)
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

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
