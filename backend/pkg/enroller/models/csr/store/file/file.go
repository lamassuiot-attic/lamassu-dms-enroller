package files

import (
	"enroller/pkg/enroller/models/csr/store"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
)

type File struct {
	dirPath string
	logger  log.Logger
}

func NewFile(dirPath string, logger log.Logger) store.File {
	return &File{dirPath: dirPath, logger: logger}
}

const (
	csrPerm = 0444
)

func (f *File) Insert(id int, data []byte) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, csrPerm)
	if err != nil {
		f.logger.Log("err", err, "msg", "Could not insert CSR in filesystem")
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		f.logger.Log("err", err, "msg", "Error encoding bytes as a CSR")
		os.Remove(name)
		return err
	}
	return nil
}

func (f *File) SelectByID(id int) ([]byte, error) {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		f.logger.Log("err", err, "msg", "Could not obtain CSR from filesystem")
		return nil, err
	}
	return data, nil
}

func (f *File) Delete(id int) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	err := os.Remove(name)
	if err != nil {
		f.logger.Log("err", err, "msg", "Could not delete CSR from filesystem")
		return err
	}
	return nil
}
