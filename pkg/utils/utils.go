package utils

import (
	"os"
)

func DirectoryExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DirectoriesExist(paths []string) error {
	for _, path := range paths {
		if path == "" {
			return EmptyDirectory
		}
		if !DirectoryExists(path) {
			return InvalidDirectory(path)
		}
	}

	return nil
}

func GetFilesInDirectory(path string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, path+"/"+entry.Name())
		}
	}

	return files, nil
}
