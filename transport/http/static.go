package http

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/nicolasparada/go-passwordless-demo/web"
)

func (h *handler) staticHandler() http.Handler {
	root, err := fs.Sub(web.Files, "static")
	if err != nil {
		h.logger.Printf("could not embed static files: %v\n", err)
		os.Exit(1)
	}

	return http.FileServer(&spaFileSystem{root: http.FS(root)})
}

type spaFileSystem struct {
	root http.FileSystem
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}

	return f, err
}
