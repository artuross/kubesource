package config

import (
	"errors"
	"fmt"
	"path"

	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/afero"
)

// Config represents the kubesource.yaml configuration file.
type Config struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	SourceDir  string   `yaml:"sourceDir"`
	Targets    []Target `yaml:"targets"`
}

// Target represents a target directory where rendered manifests should be saved.
type Target struct {
	Directory string  `yaml:"directory"`
	Filter    *Filter `yaml:"filter,omitempty"`
}

// Filter represents filtering options for a target.
// An empty include list means to include everything.
// An empty exclude list means to exclude nothing.
type Filter struct {
	Include []Selector `yaml:"include,omitempty"`
	Exclude []Selector `yaml:"exclude,omitempty"`
}

// Selector represents a resource selector for filtering. All non-empty fields must match.
type Selector struct {
	Kind       string            `yaml:"kind,omitempty"`
	APIVersion string            `yaml:"apiVersion,omitempty"`
	Metadata   *MetadataSelector `yaml:"metadata,omitempty"`
}

// MetadataSelector represents metadata-based filtering.
type MetadataSelector struct {
	Name      string            `yaml:"name,omitempty"`
	Namespace string            `yaml:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

// LoadConfig loads and parses a kubesource.yaml file from the specified directory using afero.Fs
func LoadConfig(afs afero.Fs, dir string) (*Config, error) {
	configPath := path.Join(dir, "kubesource.yaml")

	data, err := afero.ReadFile(afs, configPath)
	if err != nil {
		return nil, fmt.Errorf("reading kubesource.yaml from %s: %w", dir, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing kubesource.yaml from %s: %w", dir, err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("validating kubesource.yaml from %s: %w", dir, err)
	}

	return &config, nil
}

func validateConfig(config Config) error {
	if config.SourceDir == "" {
		return errors.New("sourceDir is required")
	}

	if len(config.Targets) == 0 {
		return errors.New("at least one target is required")
	}

	return nil
}
