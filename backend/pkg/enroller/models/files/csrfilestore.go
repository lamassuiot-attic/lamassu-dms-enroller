package files

import (
	"bufio"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewFile(dirPath string, caPath string, caCertFile string, caKeyFile string) *File {
	return &File{dirPath: dirPath, caPath: caPath, caCertFile: caCertFile, caKeyFile: caKeyFile}
}

type File struct {
	dirPath    string
	caPath     string
	caCertFile string
	caKeyFile  string
}

const (
	certPerm = 0444
)

func (f *File) getCsrCrtPath(id string, ext string) string {
	return f.dirPath + "/" + id + ext
}

func (f *File) getCAPath(name string) string {
	return f.caPath + "/" + name
}

func (f *File) InsertFileCSR(id int, rawData []byte) error {
	name := f.getCsrCrtPath(strconv.Itoa(id), ".csr")
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, certPerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(rawData); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func (f *File) SelectFileByID(id int) ([]byte, error) {
	name := f.getCsrCrtPath(strconv.Itoa(id), ".csr")
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) DeleteFile(id int) error {
	name := f.getCsrCrtPath(strconv.Itoa(id), ".csr")
	err := os.Remove(name)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) Serial() (*big.Int, error) {
	name := f.dirPath + "/serial"
	s := big.NewInt(2)
	if _, err := os.Stat(name); err != nil {
		if err := f.writeSerial(s); err != nil {
			return nil, err
		}
		return s, nil
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	data, err := r.ReadString('\r')
	if err != nil && err != io.EOF {
		return nil, err
	}
	data = strings.TrimSuffix(data, "\r")
	data = strings.TrimSuffix(data, "\n")
	serial, ok := s.SetString(data, 16)
	if !ok {
		return nil, errors.New("could not convert " + string(data) + " to serial number")
	}
	return serial, nil
}

func (f *File) writeSerial(serial *big.Int) error {
	name := f.dirPath + "/serial"
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("%x\n", serial.Bytes())); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func (f *File) LoadCACert() ([]byte, error) {
	name := f.getCAPath(f.caCertFile)
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) LoadCAKey() ([]byte, error) {
	name := f.getCAPath(f.caKeyFile)
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) InsertFileCert(id int, rawData []byte) error {
	name := f.getCsrCrtPath(strconv.Itoa(id), ".crt")
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, certPerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: rawData}); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func (f *File) LoadCert(id int) ([]byte, error) {
	name := f.getCsrCrtPath(strconv.Itoa(id), ".crt")
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//Test function
func (f *File) EmptyDirectory() error {
	d, err := os.Open(f.dirPath)
	if err != nil {
		return err
	}
	defer d.Close()
	err = deleteAllContent(f.dirPath, d)
	if err != nil {
		return err
	}
	d, err = os.Open(f.caPath)
	if err != nil {
		return err
	}
	defer d.Close()
	err = deleteAllContent(f.caPath, d)
	if err != nil {
		return err
	}
	return nil
}

// Test function
func deleteAllContent(path string, d *os.File) error {
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil

}
