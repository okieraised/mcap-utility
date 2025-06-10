package utils

import (
	"fmt"
	"mcap-utility/internal/constants"
	"os"
	"path/filepath"
	"strings"
)

func RemoveDirectory(objectPath string) error {
	err := os.RemoveAll(objectPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

func IsDirExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

func CreateDir(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return nil
}

func HasAnySuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix)) {
			return true
		}
	}
	return false
}

func RemoveEmptyStrings(input []string) []string {
	var result []string
	for _, str := range input {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

func ListMCAPFilesInDirectory(path string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), constants.MCAPFIleExtension) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func IsPathDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

func IsSameDirectory(p1, p2 string) (bool, error) {
	absPath1, err := filepath.Abs(p1)
	if err != nil {
		return false, err
	}

	absPath2, err := filepath.Abs(p2)
	if err != nil {
		return false, err
	}

	return absPath1 == absPath2, nil
}
