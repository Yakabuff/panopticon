package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	a := App{db: dbClient{db}}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/boards", a.fetchBoards)
		// /board?name=g&count=100&after=1231312&sort=asc&before=123131
		r.Get("/board", a.fetchBoardThreads)
		// /g/thread?id=123
		r.Get("/{board}/thread", a.fetchThread)
	})
	http.ListenAndServe(":3000", r)
}

func (a *App) fetchBoards(w http.ResponseWriter, r *http.Request) {
	b, err := a.db.getBoards()
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}

func (a *App) fetchBoardThreads(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")
	sort := r.URL.Query().Get("sort")
	count := r.URL.Query().Get("count")
	after2, err2 := strconv.ParseInt(after, 10, 64)

	if name == "" {
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
	b, err := a.db.getThreads(after2, before2, count2, name, sort)
	if err != nil {
		fmt.Println(err)
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}

func (a *App) fetchThread(w http.ResponseWriter, r *http.Request) {
	fmt.Println("thread")
	id := r.URL.Query().Get("id")
	fmt.Println(id)
}

type App struct {
	db dbClient
}
