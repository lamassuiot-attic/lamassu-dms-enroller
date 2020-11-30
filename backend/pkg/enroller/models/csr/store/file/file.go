package files

import (
	"enroller/pkg/enroller/models/csr/store"
	"io/ioutil"
	"os"
	"strconv"
)

type File struct {
	dirPath string
}

func NewFile(dirPath string) store.File {
	return &File{dirPath: dirPath}
}

const (
	csrPerm = 0444
)

func (f *File) Insert(id int, data []byte) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, csrPerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func (f *File) SelectByID(id int) ([]byte, error) {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) Delete(id int) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	err := os.Remove(name)
	if err != nil {
		return err
	}
	return nil
}
