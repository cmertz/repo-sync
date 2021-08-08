// nolint: testpackage
// testing an unexported function
package sync

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLocal_ensureParentPathExists(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "local")
	if err != nil {
		t.Error(err)
	}

	parent := tmp + string(os.PathSeparator) + "a"
	l := Local(parent + string(os.PathSeparator) + "b")

	err = l.ensureParentPathExists()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	defer os.RemoveAll(tmp)

	if _, err := os.Stat(parent); os.IsNotExist(err) {
		t.Errorf("expected path %s to exist", parent)
	}

	err = l.ensureParentPathExists()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	if _, err := os.Stat(parent); os.IsNotExist(err) {
		t.Errorf("expected path %s to still exist", parent)
	}
}
