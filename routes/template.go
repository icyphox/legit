package routes

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"git.icyphox.sh/legit/git"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func (d *deps) Write404(w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))
	w.WriteHeader(404)
	if err := t.ExecuteTemplate(w, "404", nil); err != nil {
		log.Printf("404 template: %s", err)
	}
}

func (d *deps) Write500(w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))
	w.WriteHeader(500)
	if err := t.ExecuteTemplate(w, "500", nil); err != nil {
		log.Printf("500 template: %s", err)
	}
}

func (d *deps) listFiles(files []git.NiceTree, data map[string]any, w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

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

func doHighlight(filename string, content string, d *deps) (string , bool){
	
	if d.c.Highlight.Enable{
		var style *chroma.Style
		
		if d.c.Highlight.Style == "" {
			style = styles.Get("github")
		} else {
			style = styles.Get(d.c.Highlight.Style)
		}

		if style == nil{
			style = styles.Fallback
		}

		lexer := lexers.Match(filename)
		if lexer == nil{
			lexer = lexers.Fallback
		}

		
		formatter := html.New(
			html.WithClasses(true),
			html.WithLineNumbers(true),
			html.WithLinkableLineNumbers(true, "L"),
		)

		itr , err := lexer.Tokenise(nil , content)
		if err != nil{
			return content, false		
		}
		buff := new(bytes.Buffer)
		_ , err = buff.WriteString("<style>")
		
		if err != nil{
			return content, false
		}

		formatter.WriteCSS(buff , style)

		_ , err = buff.WriteString("</style>")

		if err != nil{
			return content, false
		}

		err = formatter.Format(buff , style , itr)

		if err != nil{
			return content, false
		}
		return buff.String(), true
	}
		return content , false
}

func (d *deps) showFile(content string, data map[string]any, w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

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

	newContent , hasHighlight := doHighlight(
		filepath.Base(data["path"].(string)),
		content,
		d,
	)
	if hasHighlight {
		data["content"] = template.HTML(newContent)
		data["highlight"] = true 
	}else {
		data["content"] = content
		data["highlight"] = false
	}

	data["linecount"] = lines
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "file", data); err != nil {
		log.Println(err)
		return
	}
}
