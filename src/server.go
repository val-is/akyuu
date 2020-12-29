package akyuu

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-martini/martini"
)

type Akyuu struct {
	FsClient FsClient
}

func (a *Akyuu) receiveFile(w http.ResponseWriter, r *http.Request, fileTypeEndpoint fileType) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	var fType fileType
	switch header.Header.Get("Content-Type") {
	case "image/jpeg", "image/png":
		fType = fileTypeImage
	case "image/gif":
		fType = fileTypeGif
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if fType != fileTypeEndpoint {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	io.Copy(&buf, file)
	uid := genFileUID()
	fileObj := fileObject{
		UID:      uid,
		BasePath: filepath.Join(a.FsClient.StorageDir, fmt.Sprint(fType)),
		Type:     fType,
		Filename: string(uid) + header.Filename,
	}
	if err := a.FsClient.writeFile(fileObj, buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprint(fileObj.UID))
}

func (a *Akyuu) receiveImage(w http.ResponseWriter, r *http.Request) {
	a.receiveFile(w, r, fileTypeImage)
}

func (a *Akyuu) getFile(w http.ResponseWriter, r *http.Request, params martini.Params, fileTypeEndpoint fileType) {
	uid := strings.TrimSuffix(params["id"], filepath.Ext(params["id"]))
	file, present := a.FsClient.getFile(fileUID(uid))
	if !present || file.Type != fileTypeEndpoint {
		http.NotFound(w, r)
		return
	}
	if err := file.readIntoWriter(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *Akyuu) getImage(w http.ResponseWriter, r *http.Request, params martini.Params) {
	a.getFile(w, r, params, fileTypeImage)
}

func (a *Akyuu) BuildRoutes(m *martini.ClassicMartini) {
	m.Group("/i", func(r martini.Router) {
		r.Get("/:id", a.getImage)
		r.Post("", a.receiveImage)
	})
}
