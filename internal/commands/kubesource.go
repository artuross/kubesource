package commands

import (
	"context"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/afero"
	cli "github.com/urfave/cli/v3"

	"github.com/artuross/kubesource/internal/kubesource"
	"github.com/artuross/kubesource/internal/kustomize"
	"github.com/artuross/kubesource/internal/manifest"
	"github.com/artuross/kubesource/pkg/commandexec"
	"github.com/artuross/kubesource/pkg/config"
)

func NewKubesourceCommand() *cli.Command {
	return &cli.Command{
		Name:   "kubesource",
		Usage:  "vendor Kubernetes manifests from upstream sources",
		Action: runKubesourceCommand,
	}
}

func runKubesourceCommand(ctx context.Context, c *cli.Command) error {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	afs := afero.NewBasePathFs(afero.NewOsFs(), workingDir)
	executor := commandexec.NewExecutor()

	directories, err := kubesource.FindDirectories(afs, ".")
	if err != nil {
		return fmt.Errorf("finding kubesource directories: %w", err)
	}

	for _, dir := range directories {
		if err := processSingleDirectory(afs, executor, dir); err != nil {
			return fmt.Errorf("processing directory %s: %w", dir, err)
		}
	}

	return nil
}

func processSingleDirectory(afs afero.Fs, executor commandexec.CommandExecutor, baseDir string) error {
	fmt.Printf("Processing %s\n", baseDir)

	cfg, err := config.LoadConfig(afs, baseDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	for _, source := range cfg.Sources {
		sourceDir := path.Join(baseDir, source.SourceDir)

		fmt.Printf("  Source directory: %s\n", sourceDir)

		if err := kustomize.VerifyHasKustomizationFile(afs, sourceDir); err != nil {
			return fmt.Errorf("validating source directory: %w", err)
		}

		kustomizeDocument, err := kustomize.Build(executor, sourceDir)
		if err != nil {
			return fmt.Errorf("building manifests: %w", err)
		}

		parsedDocuments, err := manifest.ParseDocuments(kustomizeDocument)
		if err != nil {
			return fmt.Errorf("parsing YAML documents: %w", err)
		}

		// save to each target directory with filtering
		for _, target := range source.Targets {
			targetPath := filepath.Join(baseDir, target.Directory)

			if err := afs.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("cleaning target directory %s: %w", targetPath, err)
			}

			includedFiles, err := getTargetDocuments(parsedDocuments, target.Filter)
			if err != nil {
				return fmt.Errorf("generating target documents: %w", err)
			}

			fmt.Printf("  Saving to: %s\n", targetPath)

			if err := saveFiles(afs, targetPath, includedFiles); err != nil {
				return fmt.Errorf("saving to %s: %w", targetPath, err)
			}
		}
	}

	fmt.Printf("  ✓ Successfully processed %s\n", baseDir)

	return nil
}

func generateFilename(metadata config.Selector) string {
	kind := metadata.Kind
	name := metadata.Metadata.Name
	namespace := metadata.Metadata.Namespace

	if namespace == "" {
		return fmt.Sprintf("%s--%s.yaml", kind, name)
	}

	return fmt.Sprintf("%s--%s--%s.yaml", kind, namespace, name)
}

func getTargetDocuments(documents []manifest.ParsedDocument, filters *config.Filter) (map[string][]byte, error) {
	includedFiles := make(map[string][]byte, 0)
	for _, pd := range documents {
		if !pd.Matches(filters) {
			continue
		}

		fileName := generateFilename(pd.Metadata)
		documentContent, err := yaml.Marshal(pd.Document)
		if err != nil {
			return nil, fmt.Errorf("marshaling document to YAML: %w", err)
		}

		includedFiles[fileName] = documentContent
	}

	kustomizationPath, kustomizationData, err := kustomize.GenerateKustomizationFile(maps.Keys(includedFiles))
	if err != nil {
		return nil, fmt.Errorf("generating kustomization.yaml content: %w", err)
	}

	includedFiles[kustomizationPath] = kustomizationData

	return includedFiles, nil
}

func saveFiles(afs afero.Fs, targetDir string, files map[string][]byte) error {
	if err := afs.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", targetDir, err)
	}

	for relPath, content := range files {
		filePath := path.Join(targetDir, relPath)
		if err := afero.WriteFile(afs, filePath, content, 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}
	}

	return nil
}
