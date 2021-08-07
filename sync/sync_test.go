package sync_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cmertz/repo-sync/sync"
)

const defaultRemoteURL = "https://github.com/cmertz/repo-sync"

func TestSync(t *testing.T) {
	// this is a simplistic and quite slow end to end test for
	// a sync, anything more elaborate (e.g. an in-memory working
	// copy and locally served git remote) would have required
	// a lot more indirection and complexity in the actual code

	tmp, err := ioutil.TempDir(os.TempDir(), "sync")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp)

	s := sync.New(defaultRemoteURL, tmp)

	err = s.Do(context.Background())
	if err != nil {
		t.Error(err)
	}

	err = s.Do(context.Background())
	if err != nil {
		t.Error(err)
	}
}
