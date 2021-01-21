package scanning

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestScanEnvironments(t *testing.T) {
	commits, err := Scan("../../../gitops-repo/", "go-demo/environments", []string{"dev", "production", "staging"})
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
