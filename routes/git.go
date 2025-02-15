package routes

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"git.icyphox.sh/legit/git/service"
	securejoin "github.com/cyphar/filepath-securejoin"
)

func (d *deps) InfoRefs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	name = filepath.Clean(name)

	repo, err := securejoin.SecureJoin(d.c.Repo.ScanPath, name)
	if err != nil {
		log.Printf("securejoin error: %v", err)
		d.Write404(w)
		return
	}

	w.Header().Set("content-type", "application/x-git-upload-pack-advertisement")
	w.WriteHeader(http.StatusOK)

	cmd := service.ServiceCommand{
		Dir:    repo,
		Stdout: w,
	}

	if err := cmd.InfoRefs(); err != nil {
		http.Error(w, err.Error(), 500)
		log.Printf("git: failed to execute git-upload-pack (info/refs) %s", err)
		return
	}
}

func (d *deps) UploadPack(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	name = filepath.Clean(name)

	repo, err := securejoin.SecureJoin(d.c.Repo.ScanPath, name)
	if err != nil {
		log.Printf("securejoin error: %v", err)
		d.Write404(w)
		return
	}

	w.Header().Set("content-type", "application/x-git-upload-pack-result")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	cmd := service.ServiceCommand{
		Dir:    repo,
		Stdout: w,
	}

	var reader io.ReadCloser
	reader = r.Body

	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Printf("git: failed to create gzip reader: %s", err)
			return
		}
		defer reader.Close()
	}

	cmd.Stdin = reader
	if err := cmd.UploadPack(); err != nil {
		http.Error(w, err.Error(), 500)
		log.Printf("git: failed to execute git-upload-pack %s", err)
		return
	}
}
