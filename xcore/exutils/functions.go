package exutils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

func TempDirName(prefix string) string {
	randBytes := make([]byte, 5)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes))
}

func CreateLocalDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		println(err)
	}
}

func DeleteLocalDir(path string) {
	os.RemoveAll(path)
}

func WriteFile(path string, content string) error {
	f, cerr := os.Create(path)
	if cerr != nil {
		return cerr
	}
	_, werr := f.WriteString(content)
	if werr != nil {
		return werr
	}
	return nil
}
