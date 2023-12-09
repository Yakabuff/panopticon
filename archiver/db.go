package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type dbClient struct {
	conn *sql.DB
	// hash(board + thread id) -> set(post ids)
	cache *redisClient
}

func (d *dbClient) insertBoard(b Board) error {
	stmt := "INSERT INTO boards(board, unlisted) values($1, $2) ON CONFLICT DO NOTHING"
	_, err := d.conn.Exec(stmt, b.board, b.unlisted)
	return err
}

func (d *dbClient) insertThreadTask(tt ThreadTask) error {
	stmt := "INSERT INTO thread_backlog(board, no, last_modified, last_archived, replies, page, tid) values ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (board, no) DO UPDATE SET last_modified = $3, page = $6"
	_, err := d.conn.Exec(stmt, tt.Board, tt.No, tt.LastModified, tt.LastArchived, tt.Replies, tt.Page, tt.Tid)
	return err
}

func (d *dbClient) insertMediaTask(mt MediaTask) error {
	stmt := "INSERT INTO media_backlog(board, file, date_added, hash) values ($1, $2, $3, $4) ON CONFLICT (board, file) DO NOTHING"
	_, err := d.conn.Exec(stmt, mt.Board, mt.File, mt.DateAdded, mt.Hash)
	return err
}

func (d *dbClient) updateThreadTaskArchivedDate(tt ThreadTask) error {
	stmt := "UPDATE thread_backlog SET last_archived = $1 WHERE no = $2 and board = $3"
	_, err := d.conn.Exec(stmt, tt.LastArchived, tt.No, tt.Board)
	return err
}
func (d *dbClient) fetchMediaTask() ([]MediaTask, error) {
	var tasks []MediaTask
	stmt := "SELECT board, file, date_added, hash FROM media_backlog ORDER BY date_added ASC LIMIT 250"

	rows, err := d.conn.Query(stmt)
	if err != nil {
		fmt.Println(err)
		return tasks, err
	}
	defer rows.Close()
	for rows.Next() {
		var task MediaTask

		if err := rows.Scan(&task.Board, &task.File,
			&task.DateAdded, &task.Hash); err != nil {
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
func (d *dbClient) fetchThreadTask() ([]ThreadTask, error) {
	var tasks []ThreadTask
	now := time.Now()
	stmt := "SELECT no, board, last_modified, last_archived, replies, page, tid FROM thread_backlog where last_modified > last_archived AND last_archived < $1 ORDER BY page DESC LIMIT 250"
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
			&task.LastArchived, &task.Replies, &task.Page, &task.Tid); err != nil {
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
	d.cache.deleteTid(tt.Board, tt.Tid)
	return nil
}

func (d *dbClient) deleteMediaTask(mt MediaTask) error {
	fmt.Printf("Pruning media task file: %s board: %s\n", mt.File, mt.Board)
	stmt := "DELETE FROM media_backlog where file = $1 and board = $2"
	_, err := d.conn.Exec(stmt, mt.File, mt.Board)
	if err != nil {
		return err
	}
	return nil
}

func (d *dbClient) insertPost(board string, no int, resto int, time int, name string, trip string, com string, tid string, pid string, hasImage bool) error {
	stmt := "INSERT INTO post(no, resto, time, name, trip, com, board, tid, pid, has_image) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT DO NOTHING;"

	ok, err := d.cache.checkPidExist(board, tid, pid)
	if err != nil {
		fmt.Println(err)
	}

	// Check if pid exists in boardtid cache.  If exists, don't insert into db
	// Else insert into db and insert into cache
	if !ok {
		fmt.Println("inserting post from " + tid)
		_, err := d.conn.Exec(stmt, no, resto, time, name, trip, com, board, tid, pid, hasImage)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = d.cache.insertPid(board, tid, pid)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("skipping tid: " + tid + " pid: " + pid)
	}
	return nil
}

func (d *dbClient) insertThread(board string, no int, time int, name string, trip string, sub string, com string, replies int, images int, tid string, hasImage bool) error {
	stmt := "INSERT INTO thread(no, time, name, trip, sub, com, replies, images, board, tid, has_image) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT DO NOTHING;"
	// check if board thread key exists
	// if not, create key with set. store value 0 in set. insert into db
	// else skip
	ok, err := d.cache.checkThreadExist(board, tid)
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		_, err := d.conn.Exec(stmt, no, time, name, trip, sub, com, replies, images, board, tid, hasImage)
		if err != nil {
			return err
		}
		// Add thread to store
		err = d.cache.insertPid(board, tid, "0")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbClient) insertMedia(sha256 string, md5 string, w int, h int, fsize int, mime string) (int64, error) {
	stmt := "INSERT INTO file(sha256, md5, w, h, fsize, mime) values($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING RETURNING id;"
	var id int64
	err := d.conn.QueryRow(stmt, sha256, md5, w, h, fsize, mime).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *dbClient) insertFileMapping(filename string, no int, identifier string, ext string, board string, fileid int64, tid string, pid string) error {
	stmt := "INSERT INTO file_mapping(filename, ext, identifier, no, board, fileid, tid, pid) values($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING;"
	_, err := d.conn.Exec(stmt, filename, ext, identifier, no, board, fileid, tid, pid)
	if err != nil {
		return err
	}
	return nil
}

func (d *dbClient) updateMedia(sha256 string, mime string, md5 string, w int, h int, fsize int, target string) error {
	result := []any{}
	stmt := "UPDATE file SET"
	if sha256 != "" {
		result = append(result, sha256)
		stmt += fmt.Sprintf(" sha256 = $%d,", len(result))
	}
	if mime != "" {
		result = append(result, mime)
		stmt += fmt.Sprintf(" mime = $%d,", len(result))
	}
	if md5 != "" {
		result = append(result, md5)
		stmt += fmt.Sprintf(" md5 = $%d,", len(result))
	}
	if w != 0 {
		result = append(result, w)
		stmt += fmt.Sprintf(" w = $%d,", len(result))
	}
	if h != 0 {
		result = append(result, h)
		stmt += fmt.Sprintf(" h = $%d,", len(result))
	}
	if fsize != 0 {
		result = append(result, fsize)
		stmt += fmt.Sprintf(" fsize = $%d,", len(result))
	}
	stmt = strings.TrimSuffix(stmt, ",")
	if target == md5 {
		result = append(result, md5)
		stmt += fmt.Sprintf(" WHERE md5 = $%d", len(result))
	} else {
		result = append(result, sha256)
		stmt += fmt.Sprintf(" sha256 = $%d", len(result))
	}
	fmt.Println(stmt)
	fmt.Println(result)
	_, err := d.conn.Exec(stmt, result...)
	if err != nil {
		return err
	}
	return nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
