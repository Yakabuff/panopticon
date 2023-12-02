package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/render"
)

func (a *App) fetchBoards(w http.ResponseWriter, r *http.Request) {
	b, err := a.db.getBoards()
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}
func (a *App) serveBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := a.db.getBoards()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFS(templates, "static/boards.html")
	if err != nil {
		log.Println(err)
	}
	b := Boards{Boards: boards}
	fmt.Println(b)
	tmpl.Execute(w, b)
}
