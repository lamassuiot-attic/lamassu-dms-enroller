package files

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/lamassuiot/enroller/pkg/enroller/models/csr/store"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
		level.Error(f.logger).Log("err", err, "msg", "Could not insert CSR with ID "+strconv.Itoa(id)+" in filesystem")
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Error encoding bytas as CSR")
		os.Remove(name)
		return err
	}
	level.Info(f.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" inserted in file system")
	return nil
}

func (f *File) SelectByID(id int) ([]byte, error) {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Could not obtain CSR with ID "+strconv.Itoa(id)+" from filesystem")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" obtained from file system")
	return data, nil
}

func (f *File) Delete(id int) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".csr"
	err := os.Remove(name)
	if err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Could not delete CSR with ID "+strconv.Itoa(id)+" from filesystem")
		return err
	}
	level.Info(f.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" deleted from file system")
	return nil
}
