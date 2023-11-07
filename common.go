package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
)

type ImageBoard interface {
	fetchThread(Task, *dbClient) (any, error)
	fetchCatalog(Task) (any, error)
	fetchMedia(Task, *dbClient, *lru.Cache[string, any]) (Media, error)
	getType() ImageboardType
	threadWorker(any, *dbClient, *lru.Cache[string, any]) error
	threadWatcher(db *dbClient, h chan Task)
	mediaWatcher(db *dbClient, h chan Task)
	mediaWorker(media Media, db *dbClient)
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

type MediaTask struct {
	Board     string
	File      string
	DateAdded int64
}

type Thread struct {
	Board  string
	Thread any
}

type Media struct {
	W      int
	H      int
	Md5    string
	Sha256 string
	Fsize  int
	Mime   string
	File   string
	Board  string
}

type FileMapping struct {
	Filename string
	Ext      string
	Tim      string
	No       string
	Board    string
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
	filename string
	hash     string
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

// Write file to disk and return sha256 hash
func writeFile(reader io.Reader) (string, error) {
	path := os.Getenv("MEDIA_PATH")
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(reader)
	body2 := bytes.NewReader(body)
	//hash byte array
	sum := fmt.Sprintf("%x", sha256.Sum256(body))
	//create file with hash as file name
	newpath := filepath.Join(".", path, sum)
	_, errExist := os.Stat(newpath)
	if errExist == nil {
		//If exist, return hash and do not save file
		return sum, nil
	}

	if errors.Is(errExist, os.ErrNotExist) {
		//If file does not exist, download file and return sum
		out, err := os.Create(newpath)
		if err != nil {
			return "", err
		}
		defer out.Close()
		// Write the body to file
		_, err = io.Copy(out, body2)
		if err != nil {
			return sum, err
		}
		return sum, nil
	}
	return sum, err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
