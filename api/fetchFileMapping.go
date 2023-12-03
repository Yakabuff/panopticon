package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// type
// tid
// pid
// identifier
// filename
// ext
func (a *App) fetchFileMapping(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("type")
	if t == "thread" {
		tid := r.URL.Query().Get("tid")
		fms, err := a.db.getFileMapping(tid, "", true)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(fms)
		render.JSON(w, r, fms)
	} else if t == "post" {
		pid := r.URL.Query().Get("pid")
		tid := r.URL.Query().Get("tid")
		fms, err := a.db.getFileMapping(tid, pid, false)
		if err != nil {
			render.Status(r, 500)
			return
		}
		render.JSON(w, r, fms)
	} else {
		render.Status(r, 500)
		return
	}
}
