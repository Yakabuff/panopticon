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
	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	board := chi.URLParam(r, "board")
	var _before int64
	var _after int64
	var err error
	var hasPrev bool
	var sort string
	if before == "" && after != "" {
		// prev
		_before = 0
		_after, err = strconv.ParseInt(after, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sort = "ASC"
	} else if after == "" && before != "" {
		// next
		_before, err = strconv.ParseInt(before, 10, 64)
		hasPrev = true
		sort = "DESC"
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if after == "" && before == "" {
		// Default load catalog. no prev or next
		_before = time.Now().Unix()
		_after = 0
		hasPrev = false
		sort = "DESC"
	} else {
		// fetch range of posts?
		_before, err = strconv.ParseInt(before, 10, 64)
		hasPrev = true
		sort = "DESC"
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_after, err = strconv.ParseInt(after, 10, 64)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	ops, err := a.db.getOPs(_after, _before, 50, board, sort, "", "", "")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFS(templates, "static/catalog.html")
	if err != nil {
		log.Println(err)
	}
	// prev -> after first time sort asc limit 50 -> reverse list
	// /po?after=1231231
	// next -> before last time
	// /po?before=123123

	// Set new after and before for prev/next buttons
	if len(ops) > 0 {
		if after != "" && before == "" {
			// Coming from prev, reverse slice
			for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
				ops[i], ops[j] = ops[j], ops[i]
			}
			_after = ops[0].Time
			_before = ops[len(ops)-1].Time
		} else if after == "" && before != "" {
			// Coming from next
			_before = ops[len(ops)-1].Time
			_after = ops[0].Time
		} else if after == "" && before == "" {
			// Coming from /
			_before = ops[len(ops)-1].Time
		}
	}
	b := Ops{Ops: ops, HasPrev: hasPrev, After: _after, Before: _before}
	err = tmpl.Execute(w, b)
	if err != nil {
		fmt.Println(err)
	}
}

func (a *App) serveThread(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
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
	tmpl, err := template.ParseFS(templates, "static/thread.html")
	if err != nil {
		log.Println(err)
	}
	tmpl.Execute(w, thread)
}
