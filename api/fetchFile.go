package main

import (
	"net/http"
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
	file, err := a.db.getFileMeta(_key, sha256, md5, __w, _h, _fsize, mime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, file)
}
