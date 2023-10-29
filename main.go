package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	archiver := newArchiver()

	archiver.init()

}

type Archiver struct {
	boards               []string
	httpWorkerChannel    chan Task
	threadWorkerChannel  chan any
	threadWatcherChannel chan any
	boardWorkerChannel   chan any
	mediaWorkerChannel   chan Media
	imageboard           ImageBoard
	db                   dbClient
}

func newArchiver() *Archiver {
	b := strings.Split(os.Getenv("BOARDS"), ",")
	ib, err := newImageBoard(os.Getenv("TYPE"))
	if err != nil {
		log.Fatalln(err)
	}
	return &Archiver{boards: b, imageboard: ib}
}

func (a *Archiver) init() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	a.db = dbClient{conn: db, store: make(map[string]map[string]struct{})}
	defer db.Close()
	a.httpWorkerChannel = make(chan Task)
	a.threadWorkerChannel = make(chan any)
	a.boardWorkerChannel = make(chan any)

	go a.httpWorker()
	go a.threadWatcher()

	for w := 0; w <= 3; w++ {
		go a.threadWorker()
	}

	for _, b := range a.boards {
		go a.watchBoard(b)
	}

	select {}
}

// Send task to http worker every n seconds to fetch status of board
// Block on sending to channel to ensure wait 10 seconds
func (a *Archiver) watchBoard(board string) {
	go a.boardWorker(board)
	for {
		a.httpWorkerChannel <- Task{taskType: BOARD, board: board}
		time.Sleep(10 * time.Second)
	}
}

func (a *Archiver) watchArchive(board string) {

}

// Insert posts into db.  If new images, send download task to request thread
func (a *Archiver) threadWorker() {
	// fmt.Println("Spawned thread worker")
	for {
		t := <-a.threadWorkerChannel

		a.imageboard.threadWorker(t, &a.db)
	}

}

// Keep track of which threads are archived/deleted/newly created.
// Keep all op ids and thread watcher channels in memory (hashmap) and check if OPs are present
//
// If new thread (OP not in memory) spawn thread watcher with its own channel.
// If thread missing (deleted or archived), tell thread watcher of that thread to kill
func (a *Archiver) boardWorker(board string) {
	fmt.Println("Spawned board worker: " + board)
	a.imageboard.boardWorker(a.boardWorkerChannel, board, &a.db)
}

// Query thread tasks
func (a *Archiver) threadWatcher() {
	fmt.Println("Spawned thread watcher")
	for {
		a.imageboard.threadWatcher(&a.db, a.httpWorkerChannel)
	}

}
