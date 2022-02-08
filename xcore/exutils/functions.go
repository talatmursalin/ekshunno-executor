package exutils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

func TempDirName(prefix string) string {
	randBytes := make([]byte, 5)
	_, _ = rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes))
}

func CreateLocalDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		println(err)
	}
}

func DeleteLocalDir(path string) {
	_ = os.RemoveAll(path)
}

func WriteFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
