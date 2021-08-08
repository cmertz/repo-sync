// nolint: testpackage
// testing an unexported type
package sync

import (
	"strconv"
	"testing"
)

func TestRemote_local(t *testing.T) {
	cases := []struct {
		remote string
		prefix string
		local  string
	}{
		{
			"example.com",
			"",
			"example.com",
		},
		{
			"example.com",
			"src",
			"src/example.com",
		},
		{
			"git@example.com:user/1",
			"src",
			"src/example.com/user/1",
		},
		{
			"ssh://example.com/user/2",
			"src",
			"src/example.com/user/2",
		},
		{
			"git://example.com/user/3",
			"src",
			"src/example.com/user/3",
		},
		{
			"https://example.com/user/4",
			"src",
			"src/example.com/user/4",
		},
		{
			"http://example.com/user/5",
			"src",
			"src/example.com/user/5",
		},
	}

	for i, c := range cases {
		c := c

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			l := string(remote(c.remote).local(c.prefix))
			if l != c.local {
				t.Errorf("expected %s, actual %s", c.local, l)
			}
		})
	}
}
