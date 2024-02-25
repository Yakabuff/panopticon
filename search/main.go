package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blevesearch/bleve/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

// Thread 1:
// Get root folder
// Iterate through every folder in root
// Open indexes
// Thread 2:
// Start webserver
// Shared index pointers to write and read index at same time
// boadname -> index hashmap
// If new board, create new index
// If board exists, use index
func main() {
	indexes := make(map[string]bleve.Index)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	items, _ := ioutil.ReadDir(os.Getenv("PATH_INDEX"))
	for _, item := range items {
		if item.IsDir() {
			index, err := bleve.Open(os.Getenv("PATH_INDEX") + item.Name())
			if err != nil {
				fmt.Println("Failed to open " + item.Name())
				continue
			}
			indexes[item.Name()] = index
			fmt.Println("Opened index: " + item.Name())
		}
	}

	var i Indexer = Indexer{indexes, db}
	go i.indexDB()
	go i.startServer()
	fmt.Println("panopticon fts is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
