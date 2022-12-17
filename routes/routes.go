package routes

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexedwards/flow"
	"github.com/dustin/go-humanize"
	"icyphox.sh/legit/config"
	"icyphox.sh/legit/git"
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

	type info struct{ Name, Desc, Idle string }

	infos := []info{}

	for _, dir := range dirs {
		path := filepath.Join(d.c.Repo.ScanPath, dir.Name())
		gr, err := git.Open(path, "")
		if err != nil {
			d.Write500(w)
			log.Printf("opening dir %s: %s", path, err)
			return
		}

		c, err := gr.LastCommit()
		if err != nil {
			d.Write500(w)
			log.Println(err)
			return
		}

		var desc string
		db, err := os.ReadFile(filepath.Join(path, "description"))
		if err == nil {
			desc = string(db)
		} else {
			desc = ""
		}

		infos = append(infos, info{
			Name: dir.Name(),
			Desc: desc,
			Idle: humanize.Time(c.Author.When),
		})
	}

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
	name = filepath.Clean(name)
	path := filepath.Join(d.c.Repo.ScanPath, name)
	gr, err := git.Open(path, "")
	if err != nil {
		d.Write404(w)
		return
	}

	files, err := gr.FileTree("")
	if err != nil {
		d.Write500(w)
		log.Println(err)
		return
	}

	var readmeContent string
	for _, readme := range d.c.Repo.Readme {
		readmeContent, _ = gr.FileContent(readme)
		if readmeContent != "" {
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

	data := make(map[string]any)
	data["name"] = name
	data["ref"] = mainBranch
	data["readme"] = readmeContent

	d.listFiles(files, data, w)
	return
}

func (d *deps) RepoTree(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
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

	d.listFiles(files, data, w)
	return
}

func (d *deps) FileContent(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
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

	d.showFile(contents, data, w)
	return
}

func (d *deps) Log(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
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

	if err := t.ExecuteTemplate(w, "log", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) Diff(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
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

	if err := t.ExecuteTemplate(w, "commit", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) Refs(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")

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
