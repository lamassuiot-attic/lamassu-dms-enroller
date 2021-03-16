package file

import (
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/lamassuiot/enroller/pkg/enroller/models/certs/store"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"strconv"
)

type File struct {
	dirPath string
	logger  log.Logger
}

func NewFile(dirPath string, logger log.Logger) store.File {
	return &File{dirPath: dirPath, logger: logger}
}

const (
	certPerm = 0444
)

func (f *File) Insert(id int, data []byte) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, certPerm)
	if err != nil {
		level.Error(f.logger).Log("err", err.Error, "msg", "Could not insert certificate with ID "+strconv.Itoa(id)+" in filesystem")
		return err
	}
	defer file.Close()

	if err := pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: data}); err != nil {
		level.Error(f.logger).Log("err", err.Error, "msg", "Error encoding bytes as a certificate for certficate with ID "+strconv.Itoa(id))
		os.Remove(name)
	}
	level.Info(f.logger).Log("msg", "Certificate with ID "+strconv.Itoa(id)+" inserted in file system")

	return nil
}

func (f *File) SelectByID(id int) ([]byte, error) {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		level.Error(f.logger).Log("err", err.Error, "msg", "Could not obtain certificate with ID "+strconv.Itoa(id)+" from filesystem")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "Certificate with ID "+strconv.Itoa(id)+" obtained from file system")
	return data, nil
}

func (f *File) Delete(id int) error {
	name := f.dirPath + "/" + strconv.Itoa(id) + ".crt"
	err := os.Remove(name)
	if err != nil {
		level.Error(f.logger).Log("err", err.Error, "msg", "Could not delete certificate with ID "+strconv.Itoa(id)+" from filesystem")
		return err
	}
	level.Info(f.logger).Log("msg", "Certificate with ID "+strconv.Itoa(id)+" deleted from file system")
	return nil
}
