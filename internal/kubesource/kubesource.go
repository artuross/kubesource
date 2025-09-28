package kubesource

import (
	"fmt"
	"maps"
	"os"
	"path"
	"slices"

	"github.com/spf13/afero"
)

func FindDirectories(afs afero.Fs, dir string) ([]string, error) {
	dirs := make(map[string]struct{})

	walkFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking dir: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if path.Base(filePath) == "kubesource.yaml" {
			dirPath := path.Dir(filePath)
			dirs[dirPath] = struct{}{}
		}

		return nil
	}

	if err := afero.Walk(afs, dir, walkFunc); err != nil {
		return nil, fmt.Errorf("searching directories: %w", err)
	}

	directories := slices.Collect(maps.Keys(dirs))
	slices.Sort(directories)

	return directories, nil
}
