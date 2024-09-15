package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"
    "strings"
    "encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(1 * time.Second))

	RouteRoot(r)

	http.ListenAndServe(":3000", r)
}

func RouteRoot(r chi.Router) {
	RouteHTML(r)
	RouteStatic(r)
	RouteResume(r)
	RouteWeblog(r)
}

//go:embed embed
var EmbeddedFileSystem embed.FS

func GetEmbeddedFileSystem(sub string) fs.FS {
	fs, err := fs.Sub(EmbeddedFileSystem, fmt.Sprintf("embed/%s", sub))

	if err != nil {
		panic(err)
	}

	return fs
}

func RouteStatic(r chi.Router) {
	StaticServer := http.FileServer(http.FS(GetEmbeddedFileSystem("static")))

	r.Handle("/resume/*", StaticServer)

	r.Handle("/favicon.ico", StaticServer)
	r.Handle("/static/*", http.StripPrefix("/static", StaticServer))
}

func RouteResume(r chi.Router) {
	ResumeServer := http.FileServer(http.FS(GetEmbeddedFileSystem("resume")))

	r.Handle("/resume/*", http.StripPrefix("/resume", ResumeServer))
}

func RouteHTML(r chi.Router) {
	HTMLServer := http.FileServer(http.FS(GetEmbeddedFileSystem("html")))

	r.Handle("/*", HTMLServer)
}

func RouteWeblog(r chi.Router) {
	WeblogFileSystem := GetEmbeddedFileSystem("weblog")
	MarkdownStaticRouter(WeblogFileSystem)(r)

	// Collect the sources (.md files in the weblog root)
	sources := make([]string, 0)

	if err := fs.WalkDir(WeblogFileSystem, ".", func(p string, d fs.DirEntry, err error) error {
		if p == "." {
			return nil
		}

		sources = append(sources, p)

		return nil
	}); err != nil {
		panic(err)
	}

	// Parse the sources into articles
	articles := make([]Article, 0)

	for _, source := range sources {
		content, err := fs.ReadFile(WeblogFileSystem, source)

		if err != nil {
			panic(err)
		}

		articles = append(articles, ParseArticle(string(content)))
	}

	fmt.Printf("%v\n", len(articles))
}


type ArticleMeta struct {
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
	Slug    string   `json:"slug"`
	Summary string   `json:"summary"`
}

type Article struct {
	Meta ArticleMeta
	Body string
}

func ParseArticle(data string) Article {
	parts := strings.Split(data, "---")

	if len(parts) != 3 {
		panic("wrong matter!")
	}

	head := parts[1]
	body := parts[2]

	meta := ArticleMeta{}
	json.Unmarshal([]byte(head), &meta)

	return Article{
		Meta: meta,
		Body: body,
	}
}

func MarkdownStaticRouter(filesystem fs.FS) func(chi.Router) {
    sourcePaths := make([]string, 0)

	if err := fs.WalkDir(filesystem, ".", func(p string, d fs.DirEntry, err error) error {
		if p == "." {
			return nil
		}

		sourcePaths = append(sourcePaths, p)

		return nil
	}); err != nil {
		panic(err)
	}

    fmt.Printf("%v\n", len(sourcePaths))

	return func(chi.Router) {
	}
}
