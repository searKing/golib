package file_store

import (
	"encoding/json"
	"fmt"
	"github.com/tus/tusd"
	"github.com/tus/tusd/filestore"
	"github.com/tus/tusd/uid"
	"io/ioutil"
	"os"
	"path/filepath"
)

var defaultFilePerm = os.FileMode(0664)

type FileStore struct {
	filestore.FileStore
}

func New(path string) *FileStore {
	return &FileStore{
		FileStore: filestore.New(path),
	}
}

func (store FileStore) NewUpload(info tusd.FileInfo) (id string, err error) {
	var uploadId string
	if info.ID == "" {
		uploadId = uid.Uid()
	} else {
		// certain tests set info.ID in advance
		uploadId = info.ID
	}
	info.ID = uploadId

	// Create .bin file with no content
	file, err := os.OpenFile(store.binPath(id), os.O_CREATE|os.O_WRONLY, defaultFilePerm)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("upload directory does not exist: %s", store.Path)
		}
		return "", err
	}
	defer file.Close()

	// writeInfo creates the file by itself if necessary
	err = store.writeInfo(id, info)
	return
}

// binPath returns the path to the .bin storing the binary data.
func (store FileStore) binPath(id string) string {
	return filepath.Join(store.Path, id+".bin")
}

// infoPath returns the path to the .info file storing the file's info.
func (store FileStore) infoPath(id string) string {
	return filepath.Join(store.Path, id+".info")
}

// writeInfo updates the entire information. Everything will be overwritten.
func (store FileStore) writeInfo(id string, info tusd.FileInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(store.infoPath(id), data, defaultFilePerm)
}
