package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type YotsubaCatalogPage struct {
	Page    int                    `json:"page"`
	Threads []YotsubaCatalogThread `json:"threads"`
}

type YotsubaCatalog struct {
	Catalog []YotsubaCatalogPage
}

type YotsubaCatalogThread struct {
	No            int `json:"no"`
	Last_modified int `json:"last_modified"`
	Replies       int `json:"replies"`
}

type YotsubaThread struct {
	Posts []YotsubaPost `json:"posts"`
}

type YotsubaBoards struct {
	Boards []YotsubaBoard `json:"boards"`
}

type YotsubaBoard struct {
	Board string `json:"board"`
}

type YotsubaPost struct {
	No           int    `json:"no"`
	Resto        int    `json:"resto"`
	Time         int    `json:"time"`
	Name         string `json:"name"`
	Trip         string `json:"trip"`
	Capcode      string `json:"capcode"`
	Country      string `json:"country"`
	Country_name string `json:"country_name"`
	Board_flag   string `json:"board_flag"`
	Flag_name    string `json:"flag_name"`
	Sub          string `json:"sub"`
	Com          string `json:"com"`
	Tim          int    `json:"tim"`
	Filename     string `json:"filename"`
	Ext          string `json:"ext"`
	Md5          string `json:"md5"`
	W            int    `json:"w"`
	H            int    `json:"h"`
	Replies      int    `json:"replies"`
	Images       int    `json:"images"`
	Archived     int    `json:"archived"`
	Fsize        int    `json:"fsize"`
}

type Yotsuba struct {
	API_ROOT        string
	API_IMG         string
	ThumbnailBoards []string
	FullImageBoards []string
}

func newYotsuba() *Yotsuba {
	tb := os.Getenv("THUMBNAIL_BOARDS")
	fib := os.Getenv("FULL_IMAGE_BOARDS")
	return &Yotsuba{
		API_ROOT:        "https://a.4cdn.org",
		API_IMG:         "https://i.4cdn.org",
		ThumbnailBoards: strings.Split(tb, ","),
		FullImageBoards: strings.Split(fib, ","),
	}
}

func (y *Yotsuba) getType() ImageboardType {
	return YOTSUBA
}

func (y *Yotsuba) fetchCatalog(task Task) (any, error) {
	resp, err := http.Get(y.API_ROOT + "/" + task.board + "/threads.json")
	if err != nil || resp.StatusCode != 200 {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var result []YotsubaCatalogPage
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println(err)
		fmt.Println("Can not unmarshal JSON")
	}
	// fmt.Println(result)
	return YotsubaCatalog{Catalog: result}, nil
}

// Fetch thread from yotsuba
// If thread 404, remove from thread backlog and from posts hashmap
func (y *Yotsuba) fetchThread(task Task, db *dbClient) (any, error) {
	resp2, err := http.Get(y.API_ROOT + "/" + task.board + "/thread/" + strconv.Itoa(task.id) + ".json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		fmt.Printf("Failed to fetch %d Status: %d Board: %s", task.id, resp2.StatusCode, task.board)
		return YotsubaThread{}, ErrInvalidStatusCode
	}

	fmt.Println(resp2.Status)
	body2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		fmt.Println(err)
	}
	var result YotsubaThread
	if err := json.Unmarshal(body2, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println(err)
		fmt.Println("Can not unmarshal JSON")
		return YotsubaThread{}, err
	}
	// fmt.Println(result)
	return result, nil
}

// Download image/thumb and get sha256 hash and mime type
// Insert md5 of media into LRU cache ASAP to prevent img/thumbnail from downloading again
func (y *Yotsuba) fetchMedia(task Task, db *dbClient, lru *lru.Cache[string, any]) (Media, error) {
	isThumbnail := y.isThumbnail(task.filename)
	img, err := http.Get(y.API_IMG + "/" + task.board + "/" + task.filename)
	if err != nil {
		fmt.Println("Media req failed")
		fmt.Println(err)
		return Media{}, err
	}

	if img.StatusCode != 200 {
		fmt.Printf("Media req failed %d \n", img.StatusCode)
		return Media{File: task.filename, Board: task.board}, ErrInvalidStatusCode
	}

	defer img.Body.Close()
	var shouldWrite bool
	if isThumbnail && stringInSlice(task.board, y.ThumbnailBoards) ||
		!isThumbnail && stringInSlice(task.board, y.FullImageBoards) ||
		isThumbnail && stringInSlice(task.board, y.FullImageBoards) {
		shouldWrite = true
	} else {
		shouldWrite = false
	}
	body, err := io.ReadAll(img.Body)
	mimeType := http.DetectContentType(body)
	defer img.Body.Close()

	if err != nil {
		fmt.Println("Failed to read request body")
		fmt.Println(err)
		return Media{}, err
	}
	hash, err := writeFile(body, isThumbnail, task.hash, shouldWrite)
	if err != nil {
		fmt.Println("Failed to write file")
		fmt.Println(err)
		return Media{}, err
	}
	lru.Add(task.hash, nil)
	return Media{Sha256: hash, Mime: mimeType, Md5: task.hash, File: task.filename, Board: task.board, IsThumbnail: isThumbnail}, nil
}

// https://a.4cdn.org/boards.json
func (y *Yotsuba) fetchBoards() ([]string, error) {
	resp, err := http.Get(y.API_ROOT + "/" + "boards.json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Println("Could not fetch boards")
		return nil, errors.New("could not fetch board")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var boards YotsubaBoards
	if err := json.Unmarshal(body, &boards); err != nil { // Parse []byte to go struct pointer
		fmt.Println(err)
		fmt.Println("Can not unmarshal JSON")
		return nil, err
	}
	var res []string
	for _, t := range boards.Boards {
		res = append(res, t.Board)
	}
	// fmt.Println(result)
	return res, nil
}

func (y *Yotsuba) threadWorker(thread any, db *dbClient, lru *lru.Cache[string, any]) error {

	z := thread.(Thread)
	board := z.Board
	x := z.Thread.(YotsubaThread)

	if z == (Thread{}) {
		err := db.deleteThreadTask(ThreadTask{No: z.Id, Board: board})
		if err != nil {
			return err
		}
	}
	// Sort to get op thread ID
	sort.Slice(x.Posts, func(i, j int) bool { return x.Posts[i].No < x.Posts[j].No })
	// Calculate internal tid hash(thread number, thread time, board)
	tid := fmt.Sprintf("%x", sha256.Sum256([]byte(strconv.Itoa(x.Posts[0].No)+strconv.Itoa(x.Posts[0].Time)+board)))
	fmt.Println("threadWorker tid " + tid)
	for _, t := range x.Posts {
		if t.Resto == 0 {
			fmt.Println("inserting thread")
			err := db.insertThread(board, t.No, t.Time, t.Name, t.Trip, t.Sub, t.Com, t.Replies, t.Images, tid)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if t.Archived == 1 {
				fmt.Printf("Yotsuba: Board %s Thread %d archived; deleting thread task\n", board, t.No)
				err = db.deleteThreadTask(ThreadTask{No: t.No, Board: board})
				if err != nil {
					fmt.Println(err)
				}
			} else {
				// Update thread job after updating/inserting thread
				fmt.Printf("Yotsuba: Board %s Thread %d inserted; updating thread task\n", board, t.No)
				time := time.Now().Unix()
				err = db.updateThreadTaskArchivedDate(ThreadTask{No: t.No, Board: board, LastArchived: time})
				if err != nil {
					fmt.Println(err)
				}
			}
		} else {
			// Insert post
			fmt.Println("insert post " + tid)
			err := db.insertPost(board, t.No, t.Resto, t.Time, t.Name, t.Trip, t.Com, tid)
			if err != nil {
				fmt.Println(err)
			}
		}
		// Queue image if present in thread
		// and hash does not already exist in LRU cache and image enabled for board
		_, mediaCached := lru.Get(t.Md5)
		if t.Filename != "" && !mediaCached {
			// Insert file mapping and media info
			fileid, err := db.insertMedia("", t.Md5, t.W, t.H, t.Fsize, "")
			if err != nil {
				fmt.Println(err)
			}
			err = db.insertFileMapping(t.Filename, t.No, t.Tim, t.Ext, board, fileid)
			if err != nil {
				fmt.Println(err)
			}
			//Queue image
			err = db.insertMediaTask(MediaTask{
				Board:     board,
				DateAdded: time.Now().Unix(),
				File:      strconv.Itoa(t.Tim) + t.Ext,
				Hash:      t.Md5,
			})
			if err != nil {
				fmt.Println(err)
			}
			// if stringInSlice(board, y.FullImageBoards) {
			// 	err := db.insertMediaTask(MediaTask{
			// 		Board:     board,
			// 		DateAdded: time.Now().Unix(),
			// 		File:      strconv.Itoa(t.Tim) + t.Ext,
			// 	})
			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}
			// }
			// // Queue thumbnail
			// if stringInSlice(board, y.ThumbnailBoards) {
			// 	err := db.insertMediaTask(MediaTask{
			// 		Board:     board,
			// 		DateAdded: time.Now().Unix(),
			// 		File:      strconv.Itoa(t.Tim) + "s.jpg",
			// 	})
			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}
			// }
		}

	}
	return nil
}

// Board worker for yotsuba
// One board worker for every board
// Iterate over YotsubaCatalog
// Upsert jobs into catalog threads into thread backlog db
func (y *Yotsuba) boardWorker(bwc chan any, board string, db *dbClient) {

	fmt.Println("Spawned Yotsuba board worker: " + board)

	for {
		fmt.Println("boardWorker: waiting for catalog")
		t := <-bwc
		fmt.Println("boardWorker: received catalog data")
		//https://go.dev/ref/spec#Type_assertions
		z := t.(YotsubaCatalog)
		for _, v := range z.Catalog {
			for _, d := range v.Threads {
				err := db.insertThreadTask(ThreadTask{
					No:           d.No,
					Board:        board,
					LastModified: d.Last_modified,
					Replies:      d.Replies,
					Page:         v.Page})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		fmt.Println("boardWorker: finished processing new/modified threads")
	}
}

// Fetch tasks and send to http worker
func (y *Yotsuba) threadWatcher(db *dbClient, hc chan Task) {
	for {
		tasks, err := db.fetchThreadTask()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Println(tasks)
		for _, s := range tasks {
			hc <- Task{taskType: THREAD, board: s.Board, id: s.No}
		}
	}
}

// Fetch tasks and send to http worker
func (y *Yotsuba) mediaWatcher(db *dbClient, hc chan Task) {
	for {
		tasks, err := db.fetchMediaTask()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Println(tasks)
		for _, s := range tasks {
			hc <- Task{taskType: MEDIA, board: s.Board, filename: s.File, hash: s.Hash}
		}
	}
}

// Process image: sha256, mime
// Update file sha in file table
// Delete image task
// If thumbnail enabled, queue thumbnail
func (y *Yotsuba) mediaWorker(media Media, db *dbClient) {

	if media.Sha256 == "" && media.Mime == "" {
		err := db.deleteMediaTask(MediaTask{File: media.File, Board: media.Board})
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if !media.IsThumbnail && stringInSlice(media.Board, y.ThumbnailBoards) ||
		!media.IsThumbnail && stringInSlice(media.Board, y.FullImageBoards) {
		//Queue thumbnail
		fmt.Println("QUEUING THUMBNAIL")
		tmp := strings.Split(media.File, ".")
		filename := tmp[0] + "s.jpg"
		err := db.insertMediaTask(MediaTask{
			Board:     media.Board,
			DateAdded: time.Now().Unix(),
			File:      filename,
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	err := db.updateMedia(media.Sha256, media.Mime, media.Md5, 0, 0, 0, media.Md5)
	if err != nil {
		fmt.Println(err)
	}
	err = db.deleteMediaTask(MediaTask{File: media.File, Board: media.Board})
	if err != nil {
		fmt.Println(err)
	}

}

func (y *Yotsuba) isThumbnail(file string) bool {
	return strings.HasSuffix(file, "s.jpg")
}
