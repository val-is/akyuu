package akyuu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-martini/martini"
)

func BuildRoutes(m *martini.ClassicMartini) {
	// functions hidden behind tokens
	m.Group("/api", func(apiRouter martini.Router) {
		apiRouter.Group("/upload", func(uploadRouter martini.Router) {
			uploadRouter.Post("/i", verifyFileImageEndpoint, receiveFile)
			uploadRouter.Post("/g", verifyFileGifEndpoint, receiveFile)
			uploadRouter.Post("/v", verifyFileVideoEndpoint, receiveFile)
		}, bindIncomingFile)

		apiRouter.Group("/token", func(tokenRouter martini.Router) {
			tokenRouter.Get("/?", listTokens)
			tokenRouter.Get("/:id", getToken)
			tokenRouter.Get("/active", listActiveTokens)
			tokenRouter.Post("/:name", createToken)
			tokenRouter.Post("/deactivate/:id", deactivateToken)
			tokenRouter.Post("/op/:id", opToken)
			tokenRouter.Post("/deop/:id")
		}, verifyIssuerToken)
	}, verifyToken)

	// file serving/public access
	m.Group("/f", func(downloadRouter martini.Router) {
		downloadRouter.Get("/i/:id", getImage)
		downloadRouter.Get("/g/:id", getGif)
		downloadRouter.Get("/v/:id", getVideo)
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
	case "image/jpeg", "image/jpg", "image/png":
		fType = FileTypeImage
	case "image/gif":
		fType = FileTypeGif
	case "video/mpeg", "video/webm":
		fType = FileTypeVideo
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

func verifyFileGifEndpoint(c martini.Context, w http.ResponseWriter, fileData multipart.File, fileHeader *multipart.FileHeader) {
	verifyFileCorrectEndpoint(c, w, fileData, fileHeader, FileTypeGif)
}

func verifyFileVideoEndpoint(c martini.Context, w http.ResponseWriter, fileData multipart.File, fileHeader *multipart.FileHeader) {
	verifyFileCorrectEndpoint(c, w, fileData, fileHeader, FileTypeVideo)
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

func getGif(w http.ResponseWriter, params martini.Params, fsClient *FsClient) {
	getFile(w, params, fsClient, FileTypeGif)
}

func getVideo(w http.ResponseWriter, params martini.Params, fsClient *FsClient) {
	getFile(w, params, fsClient, FileTypeVideo)
}

func verifyToken(c martini.Context, w http.ResponseWriter, r *http.Request, tokenReg *TokenReg) {
	tokenKey := r.Header.Get(TokenHeaderKey)
	token, present := tokenReg.VerifyToken(TokenId(tokenKey))
	if !present {
		// TODO audit token failure
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !tokenReg.VerifyValidIssuer(token) {
		// TODO see above
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// TODO audit valid token
	c.Map(token)
}

func verifyIssuerToken(c martini.Context, w http.ResponseWriter, token Token, tokenReg *TokenReg) {
	if !tokenReg.VerifyIssuerPerms(token) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func createToken(w http.ResponseWriter, params martini.Params, issuerToken Token, tokenReg *TokenReg) {
	tokenName := params["name"]
	newToken, err := tokenReg.CreateToken(tokenName, issuerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprint(newToken.ID))
}

func listTokens(w http.ResponseWriter, tokenReg *TokenReg) {
	sendJson(w, tokenReg.ListTokens(false))
}

func listActiveTokens(w http.ResponseWriter, tokenReg *TokenReg) {
	sendJson(w, tokenReg.ListTokens(true))
}

func deactivateToken(w http.ResponseWriter, params martini.Params, tokenReg *TokenReg) {
	tokenId := TokenId(params["id"])
	oldToken, present := tokenReg.GetTokenById(tokenId)
	if !present {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	oldToken.Activated = false
	if err := tokenReg.UpdateToken(tokenId, oldToken); err != nil {
		// TODO logging/return error details
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getToken(w http.ResponseWriter, params martini.Params, tokenReg *TokenReg) {
	tokenId := TokenId(params["id"])
	token, present := tokenReg.GetTokenById(tokenId)
	if !present {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sendJson(w, token)
}

func opToken(w http.ResponseWriter, params martini.Params, tokenReg *TokenReg) {
	tokenId := TokenId(params["id"])
	_, present := tokenReg.GetTokenById(tokenId)
	if !present {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if added, err := tokenReg.AddIssuer(tokenId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !added {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func deopToken(w http.ResponseWriter, params martini.Params, tokenReg *TokenReg) {
	tokenId := TokenId(params["id"])
	_, present := tokenReg.GetTokenById(tokenId)
	if !present {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if deoped, err := tokenReg.RemoveIssuer(tokenId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !deoped {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func sendJson(w http.ResponseWriter, thing interface{}) {
	dat, err := json.Marshal(thing)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}
