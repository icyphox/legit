package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func AllCommits(r *git.Repository) ([]*object.Commit, error) {
	ci, err := r.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("all commits: %w", err)
	}

	commits := []*object.Commit{}
	ci.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})

	return commits, nil
}

// A nicer git tree representation.
type NiceTree struct {
	Name   string
	Mode   string
	Size   int64
	IsFile bool
}

func FilesAtHead(r *git.Repository, path string) ([]NiceTree, error) {
	head, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("getting head: %w", err)
	}

	return FilesAtRef(r, head, path)
}

func FilesAtRef(r *git.Repository, ref *plumbing.Reference, path string) ([]NiceTree, error) {
	c, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("commit object: %w", err)
	}

	files := []NiceTree{}
	tree, err := c.Tree()
	if err != nil {
		return nil, fmt.Errorf("file tree: %w", err)
	}

	if path == "" {
		files = makeNiceTree(tree.Entries)
	} else {
		o, err := tree.FindEntry(path)
		if err != nil {
			return nil, err
		}

		if !o.Mode.IsFile() {
			subtree, err := tree.Tree(path)
			if err != nil {
				return nil, err
			}

			files = makeNiceTree(subtree.Entries)
		}
	}

	return files, nil
}

func makeNiceTree(es []object.TreeEntry) []NiceTree {
	nts := []NiceTree{}
	for _, e := range es {
		mode, _ := e.Mode.ToOSFileMode()
		nts = append(nts, NiceTree{
			Name:   e.Name,
			Mode:   mode.String(),
			IsFile: e.Mode.IsFile(),
		})
	}

	return nts
}
