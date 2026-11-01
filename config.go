package main

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"

	"github.com/cmertz/repo-sync/source"
	"github.com/cmertz/repo-sync/sync"
)

// Keys of a source in the configuration file, besides the ones source.New
// reads.
const (
	keyKind   = "kind"
	keyPrefix = "prefix"
)

// readTemplates reads the configuration file at path, a YAML list of sources,
// and returns a sync.Template for each source.
func readTemplates(path string) ([]sync.Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	tpls, err := parseTemplates(b)
	if err != nil {
		return nil, fmt.Errorf("config %s: %w", path, err)
	}

	return tpls, nil
}

func parseTemplates(b []byte) ([]sync.Template, error) {
	var entries []map[string]string

	err := yaml.Unmarshal(b, &entries)
	if err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	tpls := make([]sync.Template, 0, len(entries))

	for i, e := range entries {
		s, err := source.New(e[keyKind], e)
		if err != nil {
			return nil, fmt.Errorf("source %d: %w", i+1, err)
		}

		tpls = append(tpls, sync.Template{
			RemoteSource: s,
			LocalPrefix:  e[keyPrefix],
		})
	}

	return tpls, nil
}
