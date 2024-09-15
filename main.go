package main

import (
    "net/http"
    "time"
    "embed"

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

    Root(r)

    http.ListenAndServe(":3000", r)
}

func Root(r chi.Router) {
	Static(r)
	HTML(r)
}

//go:embed favicon.ico resume style.css
var StaticFilesystem embed.FS
var StaticServer = http.FileServer(http.FS(StaticFilesystem))

func Static(r chi.Router) {
	r.Handle("/favicon.ico", StaticServer)
	r.Handle("/style.css", StaticServer)
	r.Handle("/static/*", StaticServer)
	r.Handle("/resume/*", StaticServer)
}

//go:embed index.html weblog
var HTMLFilesystem embed.FS
var HTMLServer = http.FileServer(http.FS(HTMLFilesystem))

func HTML(r chi.Router) {
    r.Handle("/*", HTMLServer)
}
