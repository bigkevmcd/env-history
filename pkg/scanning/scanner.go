package scanning

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// Scan detects the most recent commit for a change in an environments files.
//
// Scan accepts a list of environments and a base along with a cloned
// Repository, and scans back through the history looking for changes in
// commits until it's found the most recent commit for each environment.
func Scan(r *git.Repository, base string, environments []string) (map[string]string, error) {
	commitIter, err := r.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get the commit objects from repository in %v: %w", r, err)
	}
	defer commitIter.Close()

	envCommits := map[string]string{}
	err = commitIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()
		currentDirState, err := c.Tree()
		if err != nil {
			return fmt.Errorf("failed to get tree for commit %s: %w", hash[:7], err)
		}

		prevCommitObject, err := c.Parents().Next()
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to get the next parent for commit %s: %w", hash[:7], err)
			}
			files := currentDirState.Files()
			defer files.Close()

			files.ForEach(func(f *object.File) error {
				env := envName(f.Name, base)
				if env != "" && hasString(environments, env) {
					if _, ok := envCommits[env]; !ok {
						envCommits[env] = hash
					}
				}
				return nil
			})
			return nil
		}

		// TODO: what does this really mean?
		if prevCommitObject == nil {
			return nil
		}
		prevDirState, err := prevCommitObject.Tree()
		if err != nil {
			return fmt.Errorf("could not get tree from previous commit: %w", err)
		}
		changes, err := prevDirState.Diff(currentDirState)
		if err != nil {
			return fmt.Errorf("failed to get previous dir state diff: %w", err)
		}
		for _, ch := range changes {
			action, err := ch.Action()
			if err != nil {
				return fmt.Errorf("could not get the action for change %s: %w", ch, err)
			}
			if action == merkletrie.Modify {
				filename := ch.To.Name
				if isEmpty(ch.From) {
					filename = ch.From.Name
				}
				env := envName(filename, base)
				if env != "" && hasString(environments, env) {
					if _, ok := envCommits[env]; !ok {
						envCommits[env] = hash
					}
				}
			}
		}

		if len(envCommits) == len(environments) {
			return storer.ErrStop
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to scan commits: %w", err)
	}
	return envCommits, nil
}

// ChangedEnvironments detects environments that have changed in a specific
// hash, rooted at the base parameter, it returns a list of the environment
// names that were detected as changed.
//
// e.g. my-app as a base would detect changes in "my-app/dev" and
// "my-app/staging" and return "dev" and "staging" as the changed environments.
func ChangedEnvironments(r *git.Repository, base string, h plumbing.Hash) ([]string, error) {
	foundEnvs := map[string]bool{}
	commit, err := r.CommitObject(h)
	if err != nil {
		return nil, fmt.Errorf("failed to get the commit object from repository in %v: %w", r, err)
	}

	hash := commit.Hash.String()
	currentDirState, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit %s: %w", hash[:7], err)
	}

	prevCommitObject, err := commit.Parents().Next()
	if err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to get the next parent for commit %s: %w", hash[:7], err)
		}
		files := currentDirState.Files()
		defer files.Close()

		files.ForEach(func(f *object.File) error {
			env := envName(f.Name, base)
			foundEnvs[env] = true
			return nil
		})
	}

	// TODO: what does this really mean?
	if prevCommitObject == nil {
		return []string{}, nil
	}
	prevDirState, err := prevCommitObject.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get tree from previous commit: %w", err)
	}
	changes, err := prevDirState.Diff(currentDirState)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous dir state diff: %w", err)
	}
	for _, ch := range changes {
		action, err := ch.Action()
		if err != nil {
			return nil, fmt.Errorf("could not get the action for change %s: %w", ch, err)
		}
		if action == merkletrie.Modify {
			filename := ch.To.Name
			if isEmpty(ch.From) {
				filename = ch.From.Name
			}
			env := envName(filename, base)
			foundEnvs[env] = true
		}
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to scan commits: %w", err)
	}
	found := []string{}
	for k := range foundEnvs {
		found = append(found, k)
	}
	return found, nil
}

func removeEmpty(s []string) []string {
	r := []string{}
	for _, v := range s {
		if v != "" {
			r = append(r, v)
		}
	}
	return r
}

func hasString(s []string, v string) bool {
	for _, c := range s {
		if c == v {
			return true
		}
	}
	return false
}

func envName(filename, base string) string {
	if strings.HasPrefix(filename, base) {
		return removeEmpty(strings.Split(strings.TrimPrefix(filename, base), "/"))[0]
	}
	return ""
}

func isEmpty(o object.ChangeEntry) bool {
	return object.ChangeEntry{} == o
}
