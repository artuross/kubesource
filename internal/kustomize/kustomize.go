package kustomize

import (
	"fmt"
	"iter"
	"os/exec"
	"path"
	"path/filepath"
	"slices"

	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/afero"

	"github.com/artuross/kubesource/pkg/commandexec"
)

// Build builds a directory with Kustomize.
func Build(executor commandexec.CommandExecutor, sourceDir string) ([]byte, error) {
	if err := checkKustomizeAvailable(executor); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path for %s: %w", sourceDir, err)
	}

	output, err := executor.Exec("kustomize", "build", "--enable-helm", absPath)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("kustomize build failed for %s: %s\nStderr: %s", sourceDir, err, string(exitError.Stderr))
		}

		return nil, fmt.Errorf("kustomize build failed for %s: %w", sourceDir, err)
	}

	return output, nil
}

// GenerateKustomizationFile generates a kustomization.yaml file and returns its relative path and content.
func GenerateKustomizationFile(manifestFiles iter.Seq[string]) (string, []byte, error) {
	kustomizationPath := "kustomization.yaml"

	kustomization := map[string]any{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"resources":  slices.Sorted(manifestFiles),
	}

	yamlData, err := yaml.Marshal(kustomization)
	if err != nil {
		return "", nil, fmt.Errorf("marshaling kustomization.yaml: %w", err)
	}

	return kustomizationPath, yamlData, nil
}

// VerifyHasKustomizationFile checks if the source directory contains a kustomization.yaml file.
func VerifyHasKustomizationFile(afs afero.Fs, sourceDir string) error {
	kustomizationPath := path.Join(sourceDir, "kustomization.yaml")

	exists, err := afero.Exists(afs, kustomizationPath)
	if err != nil {
		return fmt.Errorf("checking if kustomization.yaml exists %s: %w", kustomizationPath, err)
	}

	if !exists {
		return fmt.Errorf("kustomization.yaml not found in source directory: %s", sourceDir)
	}

	return nil
}

// checkKustomizeAvailable checks if kustomize is available in PATH using the provided executor.
func checkKustomizeAvailable(executor commandexec.CommandExecutor) error {
	_, err := executor.LookPath("kustomize")
	if err != nil {
		return fmt.Errorf("kustomize not found in PATH: %w. Please install kustomize", err)
	}

	return nil
}
