//nolint:testpackage // tests unexported methods
package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocal_ensureParentPathExists(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "a")
	l := Local(filepath.Join(parent, "b"))

	// The second call checks that an existing parent is fine.
	for range 2 {
		err := l.ensureParentPathExists()
		if err != nil {
			t.Fatalf("ensureParentPathExists() error = %v", err)
		}

		_, err = os.Stat(parent)
		if err != nil {
			t.Fatalf("parent %s: %v", parent, err)
		}
	}
}

func TestLocal_isEmpty(t *testing.T) {
	empty := t.TempDir()
	full := t.TempDir()

	err := os.WriteFile(filepath.Join(full, "file"), nil, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
		want bool
	}{
		{"missing", filepath.Join(empty, "missing"), true},
		{"empty directory", empty, true},
		{"directory with entries", full, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Local(c.path).isEmpty()
			if err != nil {
				t.Fatalf("isEmpty() error = %v", err)
			}

			if got != c.want {
				t.Errorf("isEmpty() = %t, want %t", got, c.want)
			}
		})
	}
}
