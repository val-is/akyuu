package akyuu

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// this implementation is more or less a placeholder
// ideally we'd be using something like s3 to hold files

type fileType int

const (
	fileTypeImage fileType = iota
	fileTypeVideo
	fileTypeGif
)

type fileUID string

func genFileUID() fileUID {
	return fileUID(fmt.Sprint(time.Now().Unix()*10 + rand.Int63()%10))
}

type fileObject struct {
	UID      fileUID  `json:"file-uid"`
	BasePath string   `json:"path"`
	Type     fileType `json:"file-type"`
	Filename string   `json:"filename"`
}

func (f fileObject) readIntoWriter(w io.Writer) error {
	file, err := os.Open(f.getPath())
	if err != nil {
		return err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	_, err = r.WriteTo(w)
	if err != nil {
		return err
	}
	return nil
}

func (f fileObject) writeBuffer(buf bytes.Buffer) error {
	file, err := os.Create(f.getPath())
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}
	return nil
}

func (f fileObject) getPath() string {
	return filepath.Join(f.BasePath, f.Filename)
}

type FsClient struct {
	StorageDir  string                 `json:"storage-dir"`
	ClientPath  string                 `json:"config-path"`
	FileListing map[fileUID]fileObject `json:"file-listing"`
}

func NewFsClient(fsClientPath string) (FsClient, error) {
	client := FsClient{
		ClientPath: fsClientPath,
	}
	if err := client.loadFsListing(); err != nil {
		return FsClient{}, err
	}
	return client, nil
}

func (f FsClient) getFile(uid fileUID) (fileObject, bool) {
	if fo, ok := f.FileListing[uid]; ok {
		return fo, true
	}
	return fileObject{}, false
}

func (f *FsClient) writeFile(fo fileObject, buf bytes.Buffer) error {
	if err := fo.writeBuffer(buf); err != nil {
		return err
	}
	f.FileListing[fo.UID] = fo
	if err := f.dumpFileListing(); err != nil {
		return err
	}
	return nil
}

func (f *FsClient) loadFsListing() error {
	file, err := os.Open(f.ClientPath)
	if err != nil {
		return err
	}
	defer file.Close()
	d, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &f)
	if err != nil {
		return err
	}
	return nil
}

func (f FsClient) dumpFileListing() error {
	m, err := json.Marshal(f)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f.ClientPath, m, 0644)
	if err != nil {
		return err
	}
	return nil
}
