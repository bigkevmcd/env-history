package tags

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func Scan(r *git.Repository, filename string, cb func(s string, body []byte) error) error {
	tags, err := r.Tags()
	if err != nil {
		return fmt.Errorf("failed to get the tags from repository in %v: %w", r, err)
	}
	defer tags.Close()

	worktree, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get a worktree: %w", err)
	}

	return tags.ForEach(func(tag *plumbing.Reference) error {
		if err := worktree.Checkout(&git.CheckoutOptions{Branch: tag.Name()}); err != nil {
			return fmt.Errorf("failed to checkout %s (%s): %w", tag.Name(), tag.Hash(), err)
		}
		tagRef := parseTagRef(tag.String())
		filename := filepath.Join(tagRef.path, filename)
		f, err := worktree.Filesystem.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", filename, err)
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}
		return cb(tagRef.version, b)
	})
}

type profileTag struct {
	path    string
	version string
}

// TODO: this should return an error if the sizes don't match up?
func parseTagRef(s string) *profileTag {
	parts := strings.SplitN(strings.SplitN(s, "/", 3)[2], "/", 2)
	return &profileTag{path: parts[0], version: parts[1]}
}
