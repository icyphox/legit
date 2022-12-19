package git

import (
	"fmt"
	"log"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type TextFragment struct {
	Header string
	Lines  []gitdiff.Line
}

type Diff struct {
	Name struct {
		Old string
		New string
	}
	TextFragments []TextFragment
	IsBinary      bool
}

// A nicer git diff representation.
type NiceDiff struct {
	Commit struct {
		Message string
		Author  object.Signature
		This    string
		Parent  string
	}
	Stat struct {
		FilesChanged int
		Insertions   int
		Deletions    int
	}
	Diff []Diff
}

func (g *GitRepo) Diff() (*NiceDiff, error) {
	c, err := g.r.CommitObject(g.h)
	if err != nil {
		return nil, fmt.Errorf("commit object: %w", err)
	}

	patch := &object.Patch{}
	commitTree, err := c.Tree()
	parent := &object.Commit{}
	if err == nil {
		parentTree := &object.Tree{}
		if c.NumParents() != 0 {
			parent, err = c.Parents().Next()
			if err == nil {
				parentTree, err = parent.Tree()
				if err == nil {
					patch, err = parentTree.Patch(commitTree)
					if err != nil {
						return nil, fmt.Errorf("patch: %w", err)
					}
				}
			}
		} else {
			patch, err = parentTree.Patch(commitTree)
			if err != nil {
				return nil, fmt.Errorf("patch: %w", err)
			}
		}
	}

	diffs, _, err := gitdiff.Parse(strings.NewReader(patch.String()))
	if err != nil {
		log.Println(err)
	}

	nd := NiceDiff{}
	nd.Commit.This = c.Hash.String()

	if parent.Hash.IsZero() {
		nd.Commit.Parent = ""
	} else {
		nd.Commit.Parent = parent.Hash.String()
	}
	nd.Commit.Author = c.Author
	nd.Commit.Message = c.Message

	for _, d := range diffs {
		ndiff := Diff{}
		ndiff.Name.New = d.NewName
		ndiff.Name.Old = d.OldName
		ndiff.IsBinary = d.IsBinary

		for _, tf := range d.TextFragments {
			ndiff.TextFragments = append(ndiff.TextFragments, TextFragment{
				Header: tf.Header(),
				Lines:  tf.Lines,
			})
			for _, l := range tf.Lines {
				switch l.Op {
				case gitdiff.OpAdd:
					nd.Stat.Insertions += 1
				case gitdiff.OpDelete:
					nd.Stat.Deletions += 1
				}
			}
		}

		nd.Diff = append(nd.Diff, ndiff)
	}

	nd.Stat.FilesChanged = len(diffs)

	return &nd, nil
}
