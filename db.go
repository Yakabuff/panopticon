package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type dbClient struct {
	conn *sql.DB
	// hash(board + thread id) -> set(post ids)
	store map[string]map[string]struct{}
	mu    sync.Mutex
}

func (d *dbClient) insertThreadTask(tt ThreadTask) error {
	stmt := "INSERT INTO thread_backlog(board, no, last_modified, last_archived, replies, page) values ($1, $2, $3, $4, $5, $6) ON CONFLICT (board, no) DO UPDATE SET last_modified = $3, page = $6"
	_, err := d.conn.Exec(stmt, tt.Board, tt.No, tt.LastModified, tt.LastArchived, tt.Replies, tt.Page)
	return err
}

func (d *dbClient) updateThreadTaskArchivedDate(tt ThreadTask) error {
	stmt := "UPDATE thread_backlog SET last_archived = $1 WHERE no = $2 and board = $3"
	_, err := d.conn.Exec(stmt, tt.LastArchived, tt.No, tt.Board)
	return err
}

func (d *dbClient) fetchThreadTask() ([]ThreadTask, error) {
	var tasks []ThreadTask
	now := time.Now()
	stmt := "SELECT no, board, last_modified, last_archived, replies, page FROM thread_backlog where last_modified > last_archived AND last_archived < $1 ORDER BY page DESC LIMIT 250"
	// Fetch only threads that were archived more than 10 seconds ago
	rows, err := d.conn.Query(stmt, int(now.Unix()-10))
	if err != nil {
		fmt.Println(err)
		return tasks, err
	}
	defer rows.Close()
	for rows.Next() {
		var task ThreadTask

		if err := rows.Scan(&task.No, &task.Board, &task.LastModified,
			&task.LastArchived, &task.Replies, &task.Page); err != nil {
			return tasks, err
		}
		// fmt.Printf("Retrieved task: %d Board: %s\n", task.No, task.Board)
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return tasks, err
	}
	return tasks, nil
}

func (d *dbClient) deleteThreadTask(tt ThreadTask) error {
	fmt.Printf("Pruning thread task no: %d board: %s\n", tt.No, tt.Board)
	stmt := "DELETE FROM thread_backlog where no = $1 and board = $2"
	_, err := d.conn.Exec(stmt, tt.No, tt.Board)
	if err != nil {
		return err
	}
	// Delete all posts from thread from post store
	fmt.Printf("Deleted thread task from store no: %d board: %s\n", tt.No, tt.Board)
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.store, tt.Board+strconv.Itoa(tt.No))
	return nil
}

func (d *dbClient) insertPost(board string, no int, resto int, time int, name string, trip string, com string) error {
	stmt := "INSERT INTO post(no, resto, time, name, trip, com, board) values($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING;"
	post := strconv.Itoa(no)
	boardThread := board + strconv.Itoa(resto)
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.store[boardThread][post]
	if !ok {
		_, err := d.conn.Exec(stmt, no, resto, time, name, trip, com, board)
		if err != nil {
			return err
		}
		d.store[boardThread][post] = struct{}{}
	} else {
		fmt.Printf("Post %d board %s in store: skipping", no, board)
	}
	return nil
}

func (d *dbClient) insertThread(board string, no int, time int, name string, trip string, sub string, com string, replies int, images int) error {
	stmt := "INSERT INTO thread(no, time, name, trip, sub, com, replies, images, board) values($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT DO NOTHING;"
	boardThread := board + strconv.Itoa(no)
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.store[boardThread]
	if !ok {
		_, err := d.conn.Exec(stmt, no, time, name, trip, sub, com, replies, images, board)
		if err != nil {
			return err
		}
		// Add thread to store
		d.store[boardThread] = make(map[string]struct{})
	}
	return nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
