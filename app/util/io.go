package util

import (
	"encoding/base64"
	"io/ioutil"
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

// EncodeBase64Image для кодирования изображения в base64
func EncodeBase64Image(imagePath string) (string, error) {
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer imageFile.Close()

	imageData, err := ioutil.ReadAll(imageFile)
	if err != nil {
		return "", err
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return base64Image, nil
}
