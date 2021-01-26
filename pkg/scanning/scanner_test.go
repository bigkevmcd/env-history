package scanning

import (
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-cmp/cmp"
)

func TestScanEnvironments(t *testing.T) {
	r, want := makeTestRepo(t)
	commits, err := Scan(r, "go-demo/environments", []string{"dev", "production", "staging"})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, commits); diff != "" {
		t.Fatalf("failed to scan commits:\n%s", diff)
	}
}

func TestScanEnvironmentsOnlyFetchesForEnvironments(t *testing.T) {
	r := newTestRepo(t)
	r.addFile(t, "go-demo/environments/production/testfile.txt")
	r.addFile(t, "go-demo/environments/dev/testfile.txt")
	initial := r.commit(t, "Initial Commit")
	commits, err := Scan(r.Repository, "go-demo/environments", []string{"production"})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"production": initial,
	}
	if diff := cmp.Diff(want, commits); diff != "" {
		t.Fatalf("failed to scan commits:\n%s", diff)
	}
}

func TestScanEnvironmentsWithUnknownEnvironment(t *testing.T) {
	r := newTestRepo(t)
	r.addFile(t, "go-demo/environments/production/testfile.txt")
	r.addFile(t, "go-demo/environments/dev/testfile.txt")
	r.commit(t, "Initial Commit")
	commits, err := Scan(r.Repository, "go-demo/environments", []string{"staging"})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{}
	if diff := cmp.Diff(want, commits); diff != "" {
		t.Fatalf("failed to scan commits:\n%s", diff)
	}
}

func TestChangedEnvironments(t *testing.T) {
	r := newTestRepo(t)
	r.addFile(t, "go-demo/environments/production/testfile.txt")
	r.commit(t, "Initial Commit")
	r.addFileWithContent(t, "go-demo/environments/production/testfile.txt", "testfile1")
	initial := r.commit(t, "Update Files")
	envs, err := ChangedEnvironments(r.Repository, "go-demo/environments", plumbing.NewHash(initial))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"production"}
	if diff := cmp.Diff(want, envs); diff != "" {
		t.Fatalf("failed to scan changed environments for %q:\n%s", initial, diff)
	}
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()
	fs := memfs.New()
	r, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		t.Fatal(err)
	}
	return &testRepo{Repository: r, fs: fs}
}

type testRepo struct {
	*git.Repository
	fs billy.Filesystem
}

func (tr *testRepo) addFile(t *testing.T, name string) {
	t.Helper()
	tr.addFileWithContent(t, name, name)
}

func (tr *testRepo) addFileWithContent(t *testing.T, name, body string) {
	t.Helper()
	writeTestFile(t, tr.fs, name, body)
	w, err := tr.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Add(name); err != nil {
		t.Fatal(err)
	}
}

func (tr *testRepo) commit(t *testing.T, msg string) string {
	w, err := tr.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	c, err := w.Commit("Initial Commit", &git.CommitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return c.String()
}

func makeTestRepo(t *testing.T) (*git.Repository, map[string]string) {
	t.Helper()
	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		t.Fatal(err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, fs, "go-demo/environments/production/testfile.txt", "go-demo/environments/production/testfile.txt")
	if _, err := w.Add("go-demo/environments/production/testfile.txt"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, fs, "go-demo/environments/dev/testfile.txt", "go-demo/environments/dev/testfile.txt")
	if _, err := w.Add("go-demo/environments/dev/testfile.txt"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, fs, "go-demo/environments/staging/testfile.txt", "go-demo/environments/staging/testfile.txt")
	if _, err := w.Add("go-demo/environments/staging/testfile.txt"); err != nil {
		t.Fatal(err)
	}

	envs := map[string]string{}
	c, err := w.Commit("Initial Commit", &git.CommitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	envs["production"] = c.String()
	envs["staging"] = c.String()
	envs["dev"] = c.String()
	return repo, envs
}

func writeTestFile(t *testing.T, fs billy.Filesystem, name, body string) {
	t.Helper()
	newFile, err := fs.Create(name)
	defer newFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newFile.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
}
