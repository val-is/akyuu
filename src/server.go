package akyuu

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-martini/martini"
)

func BuildRoutes(m *martini.ClassicMartini, fsClient *FsClient) {
	m.Map(fsClient)

	m.Group("/upload", func(uploadRouter martini.Router) {
		uploadRouter.Post("/i", verifyFileImageEndpoint, receiveFile)
	}, bindIncomingFile)

	m.Group("/f", func(downloadRouter martini.Router) {
		downloadRouter.Get("/i/:id", getImage)
	})
}

func bindIncomingFile(c martini.Context, w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.Map(file)
	c.Map(header)

	c.Next()
	file.Close()
}

func verifyFileCorrectEndpoint(c martini.Context, w http.ResponseWriter, fileData multipart.File, fileHeader *multipart.FileHeader, endpointType FileType) {
	var fType FileType
	switch fileHeader.Header.Get("Content-Type") {
	case "image/jpeg", "image/png":
		fType = FileTypeImage
	}
	if fType != endpointType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.Map(endpointType)
}

func verifyFileImageEndpoint(c martini.Context, w http.ResponseWriter, fileData multipart.File, fileHeader *multipart.FileHeader) {
	verifyFileCorrectEndpoint(c, w, fileData, fileHeader, FileTypeImage)
}

func receiveFile(w http.ResponseWriter, fileData multipart.File, fileHeader *multipart.FileHeader, endpointType FileType, fsClient *FsClient) {
	var buf bytes.Buffer
	io.Copy(&buf, fileData)
	uid := GenFileUID()
	fileObj := FileObject{
		UID:      uid,
		BasePath: filepath.Join(fsClient.StorageDir, fmt.Sprint(endpointType)),
		Type:     endpointType,
		Filename: string(uid) + fileHeader.Filename,
	}
	if err := fsClient.WriteFile(fileObj, buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprint(fileObj.UID))
}

func getFile(w http.ResponseWriter, params martini.Params, fsClient *FsClient, endpointType FileType) {
	uid := strings.TrimSuffix(params["id"], filepath.Ext(params["id"]))
	file, present := fsClient.GetFile(FileUID(uid))
	if !present || file.Type != endpointType {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := file.ReadIntoWriter(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getImage(w http.ResponseWriter, params martini.Params, fsClient *FsClient) {
	getFile(w, params, fsClient, FileTypeImage)
}
