package http

import (
	"net/http"
	"os"
)

func staticHandler() http.Handler {
	return http.FileServer(&spaFileSystem{root: http.Dir("web/static")})
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
