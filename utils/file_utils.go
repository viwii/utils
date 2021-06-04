package utils

import (
	"io/ioutil"
	"os"
)

func CreatePath(filePath string) error {
	if !IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}

	return nil
}

func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func ReadFile(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func WriteFile(path string, data []byte) error {
	file, err := os.Create(path)
	defer file.Close()

	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
