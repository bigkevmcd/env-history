package scanning

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

var emptyChange object.ChangeEntry

func Scan(p, base string, environments []string) (map[string]string, error) {
	r, err := git.PlainOpen(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open the repository in %q: %w", p, err)
	}
	commitIter, err := r.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get the commit objects from repository in %q: %w", p, err)
	}
	defer commitIter.Close()

	envCommits := map[string]string{}
	err = commitIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()
		line := strings.Split(c.Message, "\n")
		log.Println(hash[:7], line[0])

		currentDirState, err := c.Tree()
		if err != nil {
			return fmt.Errorf("failed to get tree for commit %s: %w", hash[:7], err)
		}

		prevCommitObject, err := c.Parents().Next()
		if err != nil {
			if err == io.EOF {
				return currentDirState.Files().ForEach(func(f *object.File) error {
					log.Printf("KEVIN!!!! %s\n", f.Name)
					return nil
				})
			}
			return fmt.Errorf("failed to get the next parent for commit %s: %w", hash[:7], err)
		}

		if prevCommitObject == nil {
			log.Println("  has no previous commit")
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
				if ch.From != emptyChange {
					filename = ch.From.Name
				}
				log.Printf("KEVIN!!!! %s\n", filename)
			}
		}

		// fileIter, err := c.Files()
		// if err != nil {
		// 	return err
		// }
		// defer fileIter.Close()

		// return fileIter.ForEach(func(f *object.File) error {
		// 	log.Printf("file %q\n", f.Name)
		// 	return nil
		// })

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan commits: %w", err)
	}
	return envCommits, nil
}
