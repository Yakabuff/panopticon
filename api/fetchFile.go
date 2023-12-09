package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/render"
)

// Fetch file h
func (a *App) fetchFile(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	_key, err := strconv.Atoi(key)
	if key != "" && err != nil {
		render.Status(r, 500)
		return
	}
	if key == "" {
		_key = -1
	}
	sha256 := r.URL.Query().Get("sha256")
	md5 := r.URL.Query().Get("md5")
	_w := r.URL.Query().Get("w")
	__w, err := strconv.Atoi(_w)
	if _w != "" && err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h := r.URL.Query().Get("h")
	_h, err := strconv.Atoi(h)
	if h != "" && err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fsize := r.URL.Query().Get("fsize")
	_fsize, err := strconv.Atoi(fsize)
	if fsize != "" && err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mime := r.URL.Query().Get("mime")
	file, err := a.db.getFileMeta(int64(_key), sha256, md5, __w, _h, _fsize, mime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, file)
}

// tid + pid -> file_mapping -> file_id -> file -> hash
func (a *App) serveFile(w http.ResponseWriter, r *http.Request) {
	tid := r.URL.Query().Get("tid")
	pid := r.URL.Query().Get("pid")
	hash := r.URL.Query().Get("sha256")
	imageType := r.URL.Query().Get("type")
	var path string
	var err error
	var fms []FileMapping
	if hash == "" {
		if tid != "" && pid != "" {
			fms, err = a.db.getFileMapping(tid, pid, false)
		} else {
			fms, err = a.db.getFileMapping(tid, pid, true)
		}
		if len(fms) == 0 || err != nil {
			//fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fileId := fms[0].FileID
		file, err := a.db.getFileMeta(fileId, "", "", 0, 0, 0, "")
		if err != nil {
			//fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		hash = file.Sha256
		fmt.Println(hash)
		if hash == "" {
			hash = "asdf"
		}
	}

	if imageType == "thumb" {
		path = os.Getenv("THUMB_PATH")
	} else if imageType == "full" {
		path = os.Getenv("MEDIA_PATH")
	} else {
		path = os.Getenv("MEDIA_PATH")
	}
	path = filepath.FromSlash(path + "/" + hash[0:1] + hash[1:2] + "/" + hash)
	fmt.Println(path)
	http.ServeFile(w, r, path)
}
