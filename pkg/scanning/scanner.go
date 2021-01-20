package scanning

import (
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func Scan(s string) {
	log.Printf("scanning %q", s)
	r, err := git.PlainOpen(s)
	if err != nil {
		log.Fatal(err)
	}
	commitIter, err := r.CommitObjects()
	if err != nil {
		log.Fatal(err)
	}
	defer commitIter.Close()
	err = commitIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()
		line := strings.Split(c.Message, "\n")
		log.Println(hash[:7], line[0])

		currentDirState, err := c.Tree()
		if err != nil {
			return err
		}

		prevCommitObject, err := c.Parents().Next()
		if err != nil {
			return err
		}

		if prevCommitObject == nil {
			return nil
		}
		prevDirState, err := prevCommitObject.Tree()
		if err != nil {
			return err
		}
		changes, err := prevDirState.Diff(currentDirState)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("KEVIN!!! %#\n", changes)

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
}
