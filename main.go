// Command repo-sync clones and updates local working copies of all
// repositories that a user can access on GitHub, GitLab, and Gitea.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cmertz/repo-sync/sync"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	configPath := flag.String("config", "", "path to the YAML `file` that lists the sources (required)")
	concurrent := flag.Int("concurrent", runtime.GOMAXPROCS(0), "number of repositories to sync in parallel")

	flag.Parse()

	if *configPath == "" || *concurrent < 1 || flag.NArg() > 0 {
		flag.Usage()

		return exitUsage
	}

	templates, err := readTemplates(*configPath)
	if err != nil {
		log.Print(err)

		return exitFailure
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// A failing source or repository doesn't stop the others, but makes the
	// exit status nonzero. sync.Run calls logError from this goroutine.
	failed := false
	logError := func(err error) {
		failed = true

		log.Print(err)
	}

	var syncs []sync.Syncer

	for i, t := range templates {
		s, err := t.Syncs(ctx)
		if err != nil {
			logError(fmt.Errorf("source %d: %w", i+1, err))

			continue
		}

		syncs = append(syncs, s...)
	}

	sync.Run(ctx, syncs, *concurrent, logError)

	if ctx.Err() != nil {
		log.Print("interrupted")

		return exitFailure
	}

	if failed {
		return exitFailure
	}

	return exitOK
}
