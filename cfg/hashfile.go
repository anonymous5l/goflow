package cfg

import (
	"bytes"
	"crypto/sha1"
	// "github.com/anonymous5l/console"
	"io"
	"os"
)

type HashFile struct {
	Path string
	Hash []byte
}

func NewHashFile(path string) (*HashFile, error) {
	fh := new(HashFile)
	fh.Path = path
	if err := fh.UpdateHash(); err != nil {
		return nil, err
	}
	return fh, nil
}

func (self *HashFile) GetHash() ([]byte, error) {
	file, err := os.Open(self.Path)

	if err != nil {
		// console.Err("goflow: open %s `%s` failed!", v.Name, v.Path)
		return nil, err
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		// console.Err("goflow: calc hash %s `%s` failed!", v.Name, v.Path)
		return nil, err
	}

	return hash.Sum(nil)[:20], nil
}

func (self *HashFile) UpdateHash() error {
	if hash, err := self.GetHash(); err != nil {
		return err
	} else {
		self.Hash = hash
	}

	return nil
}

func (self *HashFile) CompareHash() (bool, error) {
	if hash, err := self.GetHash(); err != nil {
		return false, err
	} else if bytes.Compare(hash, self.Hash) == 0 {
		// console.Debug("%x %x", hash, self.Hash)
		return true, nil
	} else {
		// update last hash
		// console.Debug("%x %x", hash, self.Hash)

		self.Hash = hash
	}

	return false, nil
}
