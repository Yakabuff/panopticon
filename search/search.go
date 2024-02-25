package main

import (
	"fmt"
	"net/http"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (i Indexer) searchOp(keywords string) (*bleve.SearchResult, error) {
	query := bleve.NewMatchQuery(keywords)
	search := bleve.NewSearchRequest(query)
	search.Highlight = bleve.NewHighlight()
	searchResults, err := i.index["op"].Search(search)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(searchResults)
	return searchResults, nil
}

func (i Indexer) search(w http.ResponseWriter, r *http.Request) {

	category := r.URL.Query().Get("category")
	if i.index[category] == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if category == "op" {
		kw := r.URL.Query().Get("keywords")
		val, err := i.searchOp(kw)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte(val.String()))
	}
}

func (i Indexer) startServer() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/search", i.search)
	http.ListenAndServe(":3000", r)
}
