package git

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitRepo struct {
	r *git.Repository
	h plumbing.Hash
}

type TagList []*object.Tag

// infoWrapper wraps the property of a TreeEntry so it can export fs.FileInfo
// to tar WriteHeader
type infoWrapper struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (self TagList) Len() int {
	return len(self)
}

func (self TagList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// sorting tags in reverse chronological order
func (self TagList) Less(i, j int) bool {
	return self[i].Tagger.When.After(self[j].Tagger.When)
}

func Open(path string, ref string) (*GitRepo, error) {
	var err error
	g := GitRepo{}
	g.r, err = git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}

	if ref == "" {
		head, err := g.r.Head()
		if err != nil {
			return nil, fmt.Errorf("getting head of %s: %w", path, err)
		}
		g.h = head.Hash()
	} else {
		hash, err := g.r.ResolveRevision(plumbing.Revision(ref))
		if err != nil {
			return nil, fmt.Errorf("resolving rev %s for %s: %w", ref, path, err)
		}
		g.h = *hash
	}
	return &g, nil
}

func (g *GitRepo) Commits() ([]*object.Commit, error) {
	ci, err := g.r.Log(&git.LogOptions{From: g.h})
	if err != nil {
		return nil, fmt.Errorf("commits from ref: %w", err)
	}

	commits := []*object.Commit{}
	ci.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	return commits, nil
}

func (g *GitRepo) LastCommit() (*object.Commit, error) {
	c, err := g.r.CommitObject(g.h)
	if err != nil {
		return nil, fmt.Errorf("last commit: %w", err)
	}
	return c, nil
}

func (g *GitRepo) FileContent(path string) (string, error) {
	c, err := g.r.CommitObject(g.h)
	if err != nil {
		return "", fmt.Errorf("commit object: %w", err)
	}

	tree, err := c.Tree()
	if err != nil {
		return "", fmt.Errorf("file tree: %w", err)
	}

	file, err := tree.File(path)
	if err != nil {
		return "", err
	}

	isbin, _ := file.IsBinary()

	if !isbin {
		return file.Contents()
	} else {
		return "Not displaying binary file", nil
	}
}

func (g *GitRepo) Tags() ([]*object.Tag, error) {
	ti, err := g.r.TagObjects()
	if err != nil {
		return nil, fmt.Errorf("tag objects: %w", err)
	}

	tags := []*object.Tag{}

	_ = ti.ForEach(func(t *object.Tag) error {
		for i, existing := range tags {
			if existing.Name == t.Name {
				if t.Tagger.When.After(existing.Tagger.When) {
					tags[i] = t
				}
				return nil
			}
		}
		tags = append(tags, t)
		return nil
	})

	var tagList TagList
	tagList = tags
	sort.Sort(tagList)

	return tags, nil
}

func (g *GitRepo) Branches() ([]*plumbing.Reference, error) {
	bi, err := g.r.Branches()
	if err != nil {
		return nil, fmt.Errorf("branchs: %w", err)
	}

	branches := []*plumbing.Reference{}

	_ = bi.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref)
		return nil
	})

	return branches, nil
}

func (g *GitRepo) FindMainBranch(branches []string) (string, error) {
	for _, b := range branches {
		_, err := g.r.ResolveRevision(plumbing.Revision(b))
		if err == nil {
			return b, nil
		}
	}
	return "", fmt.Errorf("unable to find main branch")
}

// WriteTar writes itself from a tree into a binary tar file format.
// prefix is root folder to be appended.
func (g *GitRepo) WriteTar(w io.Writer, prefix string) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	c, err := g.r.CommitObject(g.h)
	if err != nil {
		return fmt.Errorf("commit object: %w", err)
	}

	tree, err := c.Tree()
	if err != nil {
		return err
	}

	walker := object.NewTreeWalker(tree, true, nil)
	defer walker.Close()

	name, entry, err := walker.Next()
	for ; err == nil; name, entry, err = walker.Next() {
		info, err := newInfoWrapper(name, prefix, &entry, tree)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := tree.File(name)
			if err != nil {
				return err
			}

			reader, err := file.Blob.Reader()
			if err != nil {
				return err
			}

			_, err = io.Copy(tw, reader)
			if err != nil {
				reader.Close()
				return err
			}
			reader.Close()
		}
	}

	return nil
}

func newInfoWrapper(
	name string,
	prefix string,
	entry *object.TreeEntry,
	tree *object.Tree,
) (*infoWrapper, error) {
	var (
		size  int64
		mode  fs.FileMode
		isDir bool
	)

	if entry.Mode.IsFile() {
		file, err := tree.TreeEntryFile(entry)
		if err != nil {
			return nil, err
		}
		mode = fs.FileMode(file.Mode)

		size, err = tree.Size(name)
		if err != nil {
			return nil, err
		}
	} else {
		isDir = true
		mode = fs.ModeDir | fs.ModePerm
	}

	fullname := path.Join(prefix, name)
	return &infoWrapper{
		name:    fullname,
		size:    size,
		mode:    mode,
		modTime: time.Unix(0, 0),
		isDir:   isDir,
	}, nil
}

func (i *infoWrapper) Name() string {
	return i.name
}

func (i *infoWrapper) Size() int64 {
	return i.size
}

func (i *infoWrapper) Mode() fs.FileMode {
	return i.mode
}

func (i *infoWrapper) ModTime() time.Time {
	return i.modTime
}

func (i *infoWrapper) IsDir() bool {
	return i.isDir
}

func (i *infoWrapper) Sys() any {
	return nil
}
