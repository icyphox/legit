package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func (g *GitRepo) FileTree(path string) ([]NiceTree, error) {
	c, err := g.r.CommitObject(g.h)
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

// A nicer git tree representation.
type NiceTree struct {
	Name      string
	Mode      string
	Size      int64
	IsFile    bool
	IsSubtree bool
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
