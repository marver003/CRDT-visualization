package config

import (
	"bytes"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

func load(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := decode(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	return nil
}

func decode(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	return decoder.Decode(out)
}

func validatePort(name string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535, got %d", name, port)
	}

	return nil
}
