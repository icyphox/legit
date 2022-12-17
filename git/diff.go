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

	var parent *object.Commit
	if len(c.ParentHashes) > 0 {
		parent, err = c.Parent(0)
		if err != nil {
			return nil, fmt.Errorf("getting parent: %w", err)
		}
	} else {
		parent = c
	}

	patch, err := parent.Patch(c)
	if err != nil {
		return nil, fmt.Errorf("patch: %w", err)
	}

	diffs, _, err := gitdiff.Parse(strings.NewReader(patch.String()))
	if err != nil {
		log.Println(err)
	}

	nd := NiceDiff{}
	nd.Commit.This = c.Hash.String()
	nd.Commit.Parent = parent.Hash.String()
	nd.Commit.Author = c.Author
	nd.Commit.Message = c.Message

	for _, d := range diffs {
		ndiff := Diff{}
		ndiff.Name.New = d.NewName
		ndiff.Name.Old = d.OldName

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
