package main

import "errors"

type ImageBoard interface {
	fetchThread(Task, *dbClient) (any, error)
	fetchCatalog(Task) (any, error)
	// fetchMedia(task Task) (interface{}, error)
	getType() ImageboardType
	threadWorker(any, *dbClient) error
	threadWatcher(db *dbClient, h chan Task)
	boardWorker(bwc chan any, board string, db *dbClient)
}

type Post struct {
	No           int
	Resto        int
	Time         int
	Name         string
	Trip         string
	Capcode      string
	Country      string
	Country_name string
	Board_flag   string
	Flag_name    string
	Sub          string
	Com          string
	Tim          int
	Filename     string
	Ext          string
	Md5          string
	W            int
	H            int
}

type CatalogThread struct {
	// Id        int
	PostCount  int
	Date       int
	LastPostId int
}

type ThreadTask struct {
	No           int
	Board        string
	LastModified int
	LastArchived int64
	Replies      int
	Page         int
}

type Thread struct {
	Board  string
	Thread any
}

type Media struct {
	No       int
	W        int
	H        int
	Filename string
	Ext      string
	Md5      string
	Sha256   string
	Tim      string
}

// type Catalog struct {
// 	Threads []CatalogThread
// }

func newImageBoard(name string) (ImageBoard, error) {
	if name == "yotsuba" {
		return newYotsuba(), nil
	}
	return nil, errors.New("invalid imageboard")
}

type Task struct {
	taskType TaskType
	board    string
	id       int
	url      string
}

type Board struct {
	board    string
	title    string
	unlisted bool
}
type TaskType int

const (
	THREAD  TaskType = iota // 0
	BOARD   TaskType = iota // 1
	MEDIA   TaskType = iota // 2
	ARCHIVE TaskType = iota // 3
)

type ImageboardType int

const (
	YOTSUBA ImageboardType = iota
)
