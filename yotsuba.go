package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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
}

type Yotsuba struct {
	API_ROOT string
}

func newYotsuba() *Yotsuba {
	return &Yotsuba{API_ROOT: "https://a.4cdn.org"}
}

func (y *Yotsuba) getType() ImageboardType {
	return YOTSUBA
}

func (y *Yotsuba) fetchCatalog(task Task) (any, error) {
	resp, err := http.Get(y.API_ROOT + "/" + task.board + "/threads.json")
	if err != nil {
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
	}
	if resp2.StatusCode != 200 {
		fmt.Printf("Failed to fetch %d Status: %d Board: %s", task.id, resp2.StatusCode, task.board)
		err := db.deleteThreadTask(ThreadTask{No: task.id, Board: task.board})
		if err != nil {
			return YotsubaThread{}, err
		}
		// Remove all posts from thread from hash store
		return YotsubaThread{}, nil
	}
	defer resp2.Body.Close()
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

func (y *Yotsuba) fetchMedia(task Task) (Media, error) {
	return Media{}, nil
}

func (y *Yotsuba) threadWorker(thread any, db *dbClient) error {

	z := thread.(Thread)
	board := z.Board
	x := z.Thread.(YotsubaThread)
	for _, t := range x.Posts {

		// Check if hash in posts

		if t.Resto == 0 {
			fmt.Println("inserting thread")
			err := db.insertThread(board, t.No, t.Time, t.Name, t.Trip, t.Sub, t.Com, t.Replies, t.Images)
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
			// fmt.Println("insert post")
			err := db.insertPost(board, t.No, t.Resto, t.Time, t.Name, t.Trip, t.Com)
			if err != nil {
				fmt.Println(err)
			}
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
				err := db.insertThreadTask(ThreadTask{No: d.No, Board: board, LastModified: d.Last_modified, Replies: d.Replies, Page: v.Page})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		fmt.Println("boardworker: finished processing new/modified threads")
	}
}

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
