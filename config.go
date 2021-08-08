package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cmertz/repo-sync/source"
	"github.com/cmertz/repo-sync/sync"

	"gopkg.in/yaml.v2"
)

const (
	keyKind   = "kind"
	keyPrefix = "prefix"
)

// TODO: split functions off of the Config type
// TODO: come up with a better name
type Config string

func (c Config) path() string {
	return string(c)
}

func (c Config) read() ([]byte, error) {
	f, err := os.Open(c.path())
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return io.ReadAll(f)
}

func (c Config) templates() ([]sync.Template, error) {
	b, err := c.read()
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	return c.templatesFromBytes(b)
}

func (Config) templatesFromBytes(b []byte) ([]sync.Template, error) {
	var entries []map[string]string

	err := yaml.Unmarshal(b, &entries)
	if err != nil {
		return nil, fmt.Errorf("templatesFromBytes: %w", err)
	}

	var tpls []sync.Template

	for _, e := range entries {
		s, err := source.New(e[keyKind], e)
		if err != nil {
			return nil, fmt.Errorf("templatesFromBytes: %w", err)
		}

		tpls = append(tpls, sync.Template{
			RemoteSource: s,
			LocalPrefix:  e[keyPrefix],
		})
	}

	return tpls, nil
}

func (c Config) syncs(ctx context.Context) ([]sync.Syncer, error) {
	tpls, err := c.templates()
	if err != nil {
		return nil, fmt.Errorf("syncs: %w", err)
	}

	var syncs []sync.Syncer

	for _, t := range tpls {
		s, err := t.Syncs(ctx)
		if err != nil {
			return nil, fmt.Errorf("syncs: %w", err)
		}

		syncs = append(syncs, s...)
	}

	return syncs, nil
}
