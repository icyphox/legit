package routes

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"git.icyphox.sh/legit/git"
)

var TmplFiles *embed.FS

func readTemplate(d *deps) *template.Template {

	if d.c.Dirs.Templates != "" {
		//read from file system
		tpath := filepath.Join(d.c.Dirs.Templates, "*")
		return template.Must(template.ParseGlob(tpath))
	} else {
		//read embedded
		tpath := filepath.Join("./templates", "*")
		return template.Must(template.ParseFS(TmplFiles, tpath))
	}
}

func (d *deps) Write404(w http.ResponseWriter) {
	t := readTemplate(d)
	w.WriteHeader(404)
	if err := t.ExecuteTemplate(w, "404", nil); err != nil {
		log.Printf("404 template: %s", err)
	}
}

func (d *deps) Write500(w http.ResponseWriter) {
	t := readTemplate(d)
	w.WriteHeader(500)
	if err := t.ExecuteTemplate(w, "500", nil); err != nil {
		log.Printf("500 template: %s", err)
	}
}

func (d *deps) listFiles(files []git.NiceTree, data map[string]any, w http.ResponseWriter) {
	t := readTemplate(d)

	data["files"] = files
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "tree", data); err != nil {
		log.Println(err)
		return
	}
}

func countLines(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	bufLen := 0
	count := 0
	nl := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if c > 0 {
			bufLen += c
		}
		count += bytes.Count(buf[:c], nl)

		switch {
		case err == io.EOF:
			/* handle last line not having a newline at the end */
			if bufLen >= 1 && buf[(bufLen-1)%(32*1024)] != '\n' {
				count++
			}
			return count, nil
		case err != nil:
			return 0, err
		}
	}
}

func (d *deps) showFile(content string, data map[string]any, w http.ResponseWriter) {
	t := readTemplate(d)

	lc, err := countLines(strings.NewReader(content))
	if err != nil {
		// Non-fatal, we'll just skip showing line numbers in the template.
		log.Printf("counting lines: %s", err)
	}

	lines := make([]int, lc)
	if lc > 0 {
		for i := range lines {
			lines[i] = i + 1
		}
	}

	data["linecount"] = lines
	data["content"] = content
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "file", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) showRaw(content string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
	return
}
