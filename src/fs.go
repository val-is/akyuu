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

type FileType int

const (
	FileTypeInvalid FileType = iota
	FileTypeImage
	FileTypeVideo
	FileTypeGif
)

type FileUID string

func GenFileUID() FileUID {
	return FileUID(fmt.Sprint(time.Now().Unix()*10 + rand.Int63()%10))
}

type FileObject struct {
	UID      FileUID  `json:"file-uid"`
	BasePath string   `json:"path"`
	Type     FileType `json:"file-type"`
	Filename string   `json:"filename"`
}

func (f FileObject) ReadIntoWriter(w io.Writer) error {
	file, err := os.Open(f.GetPath())
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

func (f FileObject) WriteBuffer(buf bytes.Buffer) error {
	file, err := os.Create(f.GetPath())
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

func (f FileObject) GetPath() string {
	return filepath.Join(f.BasePath, f.Filename)
}

type FsClient struct {
	StorageDir  string                 `json:"storage-dir"`
	ClientPath  string                 `json:"config-path"`
	FileListing map[FileUID]FileObject `json:"file-listing"`
}

func NewFsClient(fsClientPath string) (FsClient, error) {
	client := FsClient{
		ClientPath: fsClientPath,
	}
	if err := client.LoadFsListing(); err != nil {
		return FsClient{}, err
	}
	return client, nil
}

func (f FsClient) GetFile(uid FileUID) (FileObject, bool) {
	if fo, ok := f.FileListing[uid]; ok {
		return fo, true
	}
	return FileObject{}, false
}

func (f *FsClient) WriteFile(fo FileObject, buf bytes.Buffer) error {
	if err := fo.WriteBuffer(buf); err != nil {
		return err
	}
	f.FileListing[fo.UID] = fo
	if err := f.DumpFileListing(); err != nil {
		return err
	}
	return nil
}

func (f *FsClient) LoadFsListing() error {
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

func (f FsClient) DumpFileListing() error {
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
