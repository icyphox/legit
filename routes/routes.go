package routes

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.icyphox.sh/legit/config"
	"git.icyphox.sh/legit/git"
	"github.com/alexedwards/flow"
	"github.com/dustin/go-humanize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

type deps struct {
	c *config.Config
}

func (d *deps) Index(w http.ResponseWriter, r *http.Request) {
	dirs, err := os.ReadDir(d.c.Repo.ScanPath)
	if err != nil {
		d.Write500(w)
		log.Printf("reading scan path: %s", err)
		return
	}

	type info struct {
		Name, Desc, Idle string
		d                time.Time
	}

	infos := []info{}

	for _, dir := range dirs {
		if d.isIgnored(dir.Name()) {
			continue
		}

		path := filepath.Join(d.c.Repo.ScanPath, dir.Name())
		gr, err := git.Open(path, "")
		if err != nil {
			continue
		}

		c, err := gr.LastCommit()
		if err != nil {
			d.Write500(w)
			log.Println(err)
			return
		}

		desc := getDescription(path)

		infos = append(infos, info{
			Name: dir.Name(),
			Desc: desc,
			Idle: humanize.Time(c.Author.When),
			d:    c.Author.When,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[j].d.Before(infos[i].d)
	})

	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})
	data["meta"] = d.c.Meta
	data["info"] = infos

	if err := t.ExecuteTemplate(w, "index", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) RepoIndex(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}
	name = filepath.Clean(name)
	path := filepath.Join(d.c.Repo.ScanPath, name)

	gr, err := git.Open(path, "")
	if err != nil {
		d.Write404(w)
		return
	}

	commits, err := gr.Commits()
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	var readmeContent template.HTML
	for _, readme := range d.c.Repo.Readme {
		ext := filepath.Ext(readme)
		content, _ := gr.FileContent(readme)
		if len(content) > 0 {
			switch ext {
			case ".md", ".mkd", ".markdown":
				unsafe := blackfriday.Run(
					[]byte(content),
					blackfriday.WithExtensions(blackfriday.CommonExtensions),
				)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
				readmeContent = template.HTML(html)
			default:
				readmeContent = template.HTML(
					fmt.Sprintf(`<pre>%s</pre>`, content),
				)
			}
			break
		}
	}

	if readmeContent == "" {
		log.Printf("no readme found for %s", name)
	}

	mainBranch, err := gr.FindMainBranch(d.c.Repo.MainBranch)
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

	if len(commits) >= 3 {
		commits = commits[:3]
	}

	data := make(map[string]any)
	data["name"] = name
	data["ref"] = mainBranch
	data["readme"] = readmeContent
	data["commits"] = commits
	data["desc"] = getDescription(path)
	data["servername"] = d.c.Server.Name

	if err := t.ExecuteTemplate(w, "repo", data); err != nil {
		log.Println(err)
		return
	}

	return
}

func (d *deps) RepoTree(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, ref)
	if err != nil {
		d.Write404(w)
		return
	}

	files, err := gr.FileTree(treePath)
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	data := make(map[string]any)
	data["name"] = name
	data["ref"] = ref
	data["parent"] = treePath
	data["desc"] = getDescription(path)

	d.listFiles(files, data, w)
	return
}

func (d *deps) FileContent(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, ref)
	if err != nil {
		d.Write404(w)
		return
	}

	contents, err := gr.FileContent(treePath)
	data := make(map[string]any)
	data["name"] = name
	data["ref"] = ref
	data["desc"] = getDescription(path)
	data["path"] = treePath

	d.showFile(contents, data, w)
	return
}

func (d *deps) Log(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}
	ref := flow.Param(r.Context(), "ref")

	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, ref)
	if err != nil {
		d.Write404(w)
		return
	}

	commits, err := gr.Commits()
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})
	data["commits"] = commits
	data["meta"] = d.c.Meta
	data["name"] = name
	data["ref"] = ref
	data["desc"] = getDescription(path)

	if err := t.ExecuteTemplate(w, "log", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) Diff(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}
	ref := flow.Param(r.Context(), "ref")

	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, ref)
	if err != nil {
		d.Write404(w)
		return
	}

	diff, err := gr.Diff()
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})

	data["commit"] = diff.Commit
	data["stat"] = diff.Stat
	data["diff"] = diff.Diff
	data["meta"] = d.c.Meta
	data["name"] = name
	data["ref"] = ref
	data["desc"] = getDescription(path)

	if err := t.ExecuteTemplate(w, "commit", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) Refs(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	if d.isIgnored(name) {
		d.Write404(w)
		return
	}

	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, "")
	if err != nil {
		d.Write404(w)
		return
	}

	tags, err := gr.Tags()
	if err != nil {
		// Non-fatal, we *should* have at least one branch to show.
		log.Println(err)
	}

	branches, err := gr.Branches()
	if err != nil {
		log.Println(err)
		d.Write500(w)
		return
	}

	tpath := filepath.Join(d.c.Dirs.Templates, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})

	data["meta"] = d.c.Meta
	data["name"] = name
	data["branches"] = branches
	data["tags"] = tags
	data["desc"] = getDescription(path)

	if err := t.ExecuteTemplate(w, "refs", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) ServeStatic(w http.ResponseWriter, r *http.Request) {
	f := flow.Param(r.Context(), "file")
	f = filepath.Clean(filepath.Join(d.c.Dirs.Static, f))

	http.ServeFile(w, r, f)
}
