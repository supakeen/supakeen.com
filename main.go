package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

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
}

//go:embed embed
var EmbeddedFileSystem embed.FS

func GetEmbeddedFileSystem(sub string) http.FileSystem {
	fs, err := fs.Sub(EmbeddedFileSystem, fmt.Sprintf("embed/%s", sub))

	if err != nil {
		panic(err)
	}

	return http.FS(fs)
}

func RouteStatic(r chi.Router) {
	StaticServer := http.FileServer(GetEmbeddedFileSystem("static"))

	r.Handle("/favicon.ico", StaticServer)
	r.Handle("/resume/*", StaticServer)
	r.Handle("/style.css", StaticServer)

	r.Handle("/static/*", http.StripPrefix("/static", StaticServer))
}

func RouteHTML(r chi.Router) {
	HTMLServer := http.FileServer(GetEmbeddedFileSystem("html"))

	r.Handle("/*", HTMLServer)
}
