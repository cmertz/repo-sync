package sync_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cmertz/repo-sync/sync"
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

	defer os.RemoveAll(tmp)

	s := sync.Sync{
		Remote: sync.Remote(defaultRemoteURL),
		Local:  sync.Local(tmp),
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
