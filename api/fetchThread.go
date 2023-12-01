package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	count := r.URL.Query().Get("count")
	trip := r.URL.Query().Get("trip")
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
		render.Status(r, 500)
		return
	}
	count2, err2 := strconv.Atoi(count)
	if err2 != nil && count != "" {
		fmt.Println(err2)
		render.Status(r, 500)
		return
	}
	if sort != "ASC" && sort != "" && sort != "DESC" {
		render.Status(r, 500)
		return
	}
	if sort == "" {
		sort = "DESC"
	}
	if count2 > 200 {
		count2 = 200
	} else if count2 == 0 {
		count2 = 50
	}
	if before2 == 0 {
		before2 = time.Now().Unix()
	}
	b, err := a.db.getThreads(after2, before2, count2, boardName, sort, trip, hasImage)
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}
