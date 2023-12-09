package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
)

func (a *App) fetchPosts(w http.ResponseWriter, r *http.Request) {
	tid := r.URL.Query().Get("tid")
	pid := r.URL.Query().Get("pid")
	no := r.URL.Query().Get("no")
	count := r.URL.Query().Get("count")
	sort := r.URL.Query().Get("sort")
	resto := r.URL.Query().Get("resto")
	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")
	name := r.URL.Query().Get("name")
	trip := r.URL.Query().Get("trip")
	board := r.URL.Query().Get("board")
	hasImage := r.URL.Query().Get("hasImage")

	if tid != "" && pid != "" {
		b, err := a.db.getPostsByID(tid, pid)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, b)
		return
	}

	if no != "" {
		b, err := a.db.getThreadByNo(no)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, b)
		return
	}
	if resto != "" {
		_resto, err2 := strconv.ParseInt(resto, 10, 64)
		if err2 != nil {
			fmt.Println(err2)
			render.Status(r, 500)
			return
		}
		b, err := a.db.getPostByResto(_resto)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, b)
		return
	}

	if board == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if hasImage != "true" && hasImage != "false" && hasImage != "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	after2, err2 := strconv.ParseInt(after, 10, 64)
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
	b, err := a.db.getPosts(after2, before2, count2, board, sort, trip, hasImage, name)
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}
