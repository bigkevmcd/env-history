package tags

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
)

func TestScan(t *testing.T) {
	dir := tempDir(t)
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: "https://github.com/weaveworks/profiles-examples.git",
	})
	assertNoError(t, err)

	foundProfiles := []string{}
	err = Scan(r, "profile.yaml", func(tag string, body []byte) error {
		foundProfiles = append(foundProfiles, tag)
		return nil
	})
	assertNoError(t, err)

	want := []string{
		"v0.0.1",
		"v0.1.0",
		"v0.1.1",
	}
	if diff := cmp.Diff(want, foundProfiles); diff != "" {
		t.Fatalf("failed to parse profiles:\n%s", diff)
	}
}

func Test_parseTagRef(t *testing.T) {
	tagTests := []struct {
		ref  string
		want *profileTag
	}{
		{"refs/tags/bitnami-nginx/v0.0.1", &profileTag{"bitnami-nginx", "v0.0.1"}},
		{"refs/tags/weaveworks-nginx/v0.1.0", &profileTag{"weaveworks-nginx", "v0.1.0"}},
		{"refs/tags/weaveworks-nginx/v0.1.1", &profileTag{"weaveworks-nginx", "v0.1.1"}},
	}

	for _, tt := range tagTests {
		t.Run(tt.ref, func(t *testing.T) {
			tag := parseTagRef(tt.ref)
			if diff := cmp.Diff(tt.want, tag, cmp.AllowUnexported(profileTag{})); diff != "" {
				t.Fatalf("failed to parse tag:\n%s", diff)
			}
		})
	}
}

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "gnome")
	assertNoError(t, err)
	t.Cleanup(func() {
		err := os.RemoveAll(dir)
		assertNoError(t, err)
	})
	return dir
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
