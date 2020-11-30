package file

import (
	"encoding/pem"
	"enroller/pkg/enroller/models/certs/store"
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
	certPerm = 0444
)

func (f *File) Insert(id int, data []byte) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, certPerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: data}); err != nil {
		os.Remove(name)
	}
	return nil
}

func (f *File) SelectByID(id int) ([]byte, error) {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) Delete(id int) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	err := os.Remove(name)
	if err != nil {
		return err
	}
	return nil
}
