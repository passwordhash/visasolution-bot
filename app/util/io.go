package util

import (
	"os"
	"path"
)

func GetAbsolutePath(relativePath string) string {
	wd, _ := os.Getwd()
	return path.Join(wd, relativePath)
}

func WriteFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = file.Write(data)

	return err
}
