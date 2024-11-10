package util

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

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

// CreateZip создает ZIP-файл с заданными именами файлов и содержимым
func CreateZip(filenames []string, contents [][]byte, zipPath string) error {
	// Создаем новый ZIP-файл
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error creating ZIP file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	if len(filenames) != len(contents) {
		return fmt.Errorf("filenames and contents slices should have the same length")
	}

	for i, filename := range filenames {
		err = AddFileToZip(zipWriter, filename, contents[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// AddFileToZip добавляет файл с заданным содержимым в zip.Writer
func AddFileToZip(zipWriter *zip.Writer, fileName string, content []byte) error {
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file in ZIP: %v", err)
	}

	// Записываем содержимое в файл
	_, err = fileWriter.Write(content)
	if err != nil {
		return fmt.Errorf("error writing content to file in ZIP: %v", err)
	}

	return nil
}

// CreateFolder создает каталог со всеми необходимыми подпапками
func CreateFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
