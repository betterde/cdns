package spa

import (
	"embed"
	"github.com/betterde/cdns/internal/journal"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var FS embed.FS

func Serve() http.FileSystem {
	dist, err := fs.Sub(FS, "dist")
	if err != nil {
		journal.Logger.Sugar().Panicw("Error mounting front-end static resources!", err)
	}

	return http.FS(dist)
}
