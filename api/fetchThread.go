package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Fetch OP and posts
func (a *App) fetchThread(w http.ResponseWriter, r *http.Request) {
	tid := r.URL.Query().Get("tid")
	if tid == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t, err := a.db.getThreadByID(tid)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p, err := a.db.getPostsByID(tid, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	thread := Thread{Op: t, Post: p}
	render.JSON(w, r, thread)
}

func (a *App) fetchOPs(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tid := r.URL.Query().Get("tid")
	boardName := r.URL.Query().Get("board")
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")
	sort := r.URL.Query().Get("sort")
	sortBy := r.URL.Query().Get("sortby")
	count := r.URL.Query().Get("count")
	trip := r.URL.Query().Get("trip")
	name := r.URL.Query().Get("name")
	// repliesGt := r.URL.Query().Get("repliesgt")
	// repliesLt := r.URL.Query().Get("replieslt")
	// imagesGt := r.URL.Query().Get("imagesgt")
	// imagesLt := r.URL.Query().Get("imageslt")
	hasImage := r.URL.Query().Get("has_image")
	after2, err2 := strconv.ParseInt(after, 10, 64)

	if tid != "" {
		b, err := a.db.getThreadByID(tid)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, b)
		return
	}

	if id != "" {
		b, err := a.db.getThreadByNo(id)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, b)
		return
	}

	if boardName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if hasImage != "true" && hasImage != "false" && hasImage != "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err2 != nil && after != "" {
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	before2, err2 := strconv.ParseInt(before, 10, 64)
	if err2 != nil && before != "" {
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	count2, err2 := strconv.Atoi(count)
	if err2 != nil && count != "" {
		fmt.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if sort != "ASC" && sort != "" && sort != "DESC" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if sort == "" {
		sort = "DESC"
	}
	if sortBy != "" && sortBy != "replies" && sortBy != "images" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if count2 > 200 {
		count2 = 200
	} else if count2 == 0 {
		count2 = 50
	}
	if before2 == 0 {
		before2 = time.Now().Unix()
	}
	b, err := a.db.getOPs(after2, before2, count2, boardName, sort, trip, hasImage, name)
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}

func (a *App) serveCatalog(w http.ResponseWriter, r *http.Request) {
	before := time.Now().Unix()
	board := chi.URLParam(r, "board")
	ops, err := a.db.getOPs(0, before, 50, board, "DESC", "", "", "")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFS(templates, "static/catalog.html")
	if err != nil {
		log.Println(err)
	}
	b := Ops{Ops: ops}
	tmpl.Execute(w, b)
}
