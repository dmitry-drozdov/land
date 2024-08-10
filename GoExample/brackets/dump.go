package main

import (
	"os"
	"path/filepath"
)

func Dump(root string, mp map[string]string) error {
	for k, v := range mp {
		err := writeFile(root+k+".go", v)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(pathOut string, content string) error {
	err := os.MkdirAll(filepath.Dir(pathOut), 0755)
	if err != nil {
		return err
	}

	var file *os.File
	file, err = os.OpenFile(pathOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
