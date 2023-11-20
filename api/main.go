package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

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
	fmt.Println("board threads")
	name := r.URL.Query().Get("name")
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")
	sort := r.URL.Query().Get("sort")
	fmt.Println(name)
	fmt.Println(after)
	fmt.Println(before)
	fmt.Println(sort)
	b, err := a.db.getThreads(0, 0, 0, "", "")
	if err != nil {
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
