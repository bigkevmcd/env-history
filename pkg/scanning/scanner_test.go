package scanning

import (
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-cmp/cmp"
)

func TestScanEnvironments(t *testing.T) {
	r, _ := makeTestRepo(t)
	commits, err := Scan(r, "go-demo/environments", []string{"dev", "production", "staging"})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"dev":        "b776ba84f22edc22cce3277c2b0da5341f852b83",
		"staging":    "b776ba84f22edc22cce3277c2b0da5341f852b83",
		"production": "042e7239324d8c11afaf188c27ad17d7c471b770",
	}
	if diff := cmp.Diff(want, commits); diff != "" {
		t.Fatalf("failed to scan commits:\n%s", diff)
	}
}

func TestScanEnvironmentsOnlyFetchesForEnvironments(t *testing.T) {
	r, want := makeTestRepo(t)
	commits, err := Scan(r, "go-demo/environments", []string{"production"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, commits); diff != "" {
		t.Fatalf("failed to scan commits:\n%s", diff)
	}
}

func openTestRepo(t *testing.T) *git.Repository {
	r, err := git.PlainOpen("../../../gitops-repo")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func makeTestRepo(t *testing.T) (*git.Repository, map[string]string) {
	t.Helper()
	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, fs, "go-demo/environments/production/testfile.txt")
	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Add("go-demo/environments/production/testfile.txt"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, fs, "go-demo/environments/dev/testfile.txt")
	if _, err := w.Add("go-demo/environments/dev/testfile.txt"); err != nil {
		t.Fatal(err)
	}

	envs := map[string]string{}
	c, err := w.Commit("Initial Commit", &git.CommitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	envs["production"] = c.String()
	return repo, envs
}

func writeTestFile(t *testing.T, fs billy.Filesystem, name string) {
	t.Helper()
	newFile, err := fs.Create(name)
	defer newFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newFile.Write([]byte(name)); err != nil {
		t.Fatal(err)
	}
}
