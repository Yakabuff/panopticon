package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

func (a *App) fetchBoards(w http.ResponseWriter, r *http.Request) {
	b, err := a.db.getBoards()
	if err != nil {
		fmt.Println(err)
	}
	render.JSON(w, r, b)
}
