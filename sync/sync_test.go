// nolint: testpackage
// converting tho blackbox tests (i.e. `sync_test` package)
// requires exposing internals of `Sync`
package sync

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
)

const defaultRemoteURL = "https://github.com/cmertz/repo-sync"

// this is a simplistic and quite slow end to end test for
// a sync, anything more elaborate (e.g. an in-memory working
// copy and locally served git remote) would have required
// a lot more indirection and complexity in the actual code
func TestSync_Do(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "sync")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmp)

	s := Sync{
		remote: remote(defaultRemoteURL),
		local:  local(tmp),
	}

	err = s.Do(context.Background())
	if err != nil {
		t.Error(err)
	}

	err = s.Do(context.Background())
	if err != nil {
		t.Error(err)
	}
}
