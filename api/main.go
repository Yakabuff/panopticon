package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		// /op?board=g&count=100&after=1231312&sort=asc&before=123131
		r.Get("/op", a.fetchOPs)
		r.Get("/thread", a.fetchThread)
		r.Get("/post", a.fetchPosts)
		// /filemapping?id=123123&tid=123123&filename=asdfasdf&ext=asdfasdf
		r.Get("/filemapping", a.fetchFileMapping)
		r.Get("/file", a.fetchFile)
	})
	http.ListenAndServe(":3000", r)
}

type App struct {
	db dbClient
}
