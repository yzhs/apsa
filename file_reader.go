package apsa

import "io/ioutil"

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type FileReaderImpl struct{}

func (FileReaderImpl) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
