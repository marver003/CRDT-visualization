package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type LiveTest struct {
	Mode     string        `yaml:"mode"`
	Nodes    []string      `yaml:"nodes"`
	Clients  int           `yaml:"clients"`
	Requests int           `yaml:"requests"`
	Timeout  time.Duration `yaml:"timeout"`
}

func LoadLiveTest(path string) (*LiveTest, error) {
	var cfg LiveTest

	if err := load(path, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	for i, node := range cfg.Nodes {
		cfg.Nodes[i] = strings.TrimRight(node, "/")
	}

	return &cfg, nil
}

func (c *LiveTest) validate() error {
	if c.Mode != "livetest" {
		return fmt.Errorf(`mode must be "livetest", got %q`, c.Mode)
	}

	if len(c.Nodes) == 0 {
		return errors.New("no nodes defined")
	}

	for _, node := range c.Nodes {
		parsed, err := url.Parse(node)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("node %q must be a URL like http://localhost:8001", node)
		}
	}

	if c.Clients < 1 {
		return fmt.Errorf("clients must be at least 1, got %d", c.Clients)
	}

	if c.Requests < 1 {
		return fmt.Errorf("requests must be at least 1, got %d", c.Requests)
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be positive, for example 30s")
	}

	return nil
}
