package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ImageBoard interface {
	fetchThread(Task, *dbClient) (any, error)
	fetchBoards() ([]string, error)
	fetchCatalog(Task) (any, error)
	fetchMedia(Task, *dbClient, *redisClient) (Media, error)
	getType() ImageboardType
	threadWorker(any, *dbClient, *redisClient) error
	threadWatcher(db *dbClient, h chan Task)
	mediaWatcher(db *dbClient, h chan Task)
	mediaWorker(media Media, db *dbClient)
	boardWorker(bwc chan any, board string, db *dbClient)
	isThumbnail(filename string) bool
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
	Tid          string
}

type MediaTask struct {
	Board     string
	File      string
	DateAdded int64
	Hash      string
}

type Thread struct {
	Board  string
	Id     int
	Thread any
}

type Media struct {
	W           int
	H           int
	Md5         string
	Sha256      string
	Fsize       int
	Mime        string
	File        string
	Board       string
	IsThumbnail bool
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
	unlisted bool
}
type TaskType int

const (
	THREAD    TaskType = iota // 0
	BOARD     TaskType = iota // 1
	MEDIA     TaskType = iota // 2
	ARCHIVE   TaskType = iota // 3
	BOARDMETA TaskType = iota // 4
)

type ImageboardType int

const (
	YOTSUBA ImageboardType = iota
)

// Write file to disk and return sha256 hash
func writeFile(bytez []byte, isThumbnail bool, fullsizeHash string, shouldWrite bool) (string, error) {
	var path string
	if isThumbnail {
		path = os.Getenv("THUMB_PATH")
	} else {
		path = os.Getenv("MEDIA_PATH")
	}

	//hash byte array
	sum := fmt.Sprintf("%x", sha256.Sum256(bytez))

	if !shouldWrite {
		return sum, nil
	}
	var newpath string
	//create file with first 2 digits of hash as folder
	if !isThumbnail {
		newpath = filepath.Join(path, sum[0:2])
		fmt.Println(newpath)
	} else {
		newpath = filepath.Join(path, fullsizeHash[0:2])
		fmt.Println(newpath)
	}

	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		fmt.Println("failed to mkdir")
		return "", err
	}

	if isThumbnail {
		newpath = filepath.Join(newpath, fullsizeHash)
	} else {
		newpath = filepath.Join(newpath, sum)
	}
	fmt.Println(newpath)

	_, errExist := os.Stat(newpath)
	if errExist == nil {
		//If exist, return hash and do not save file
		fmt.Println("file exists " + newpath)
		return sum, nil
	}

	if errors.Is(errExist, os.ErrNotExist) {
		//If file does not exist, save file and return sum
		out, err := os.Create(newpath)
		if err != nil {
			fmt.Println("failed to create file")
			return "", err
		}
		defer out.Close()
		// Write the body to file
		body2 := bytes.NewReader(bytez)
		_, err = io.Copy(out, body2)
		if err != nil {
			fmt.Println("failed to copy content to file")
			return sum, err
		}
		return sum, nil
	}
	return sum, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

var ErrInvalidStatusCode = errors.New("invalid status code")
