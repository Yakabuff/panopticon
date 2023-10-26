package main

import (
	"database/sql"
	"fmt"
	"time"
)

type dbClient struct {
	conn *sql.DB
}

func (d *dbClient) insertThreadTask(tt ThreadTask) error {
	stmt := "INSERT INTO thread_backlog(board, no, last_modified, last_archived, replies, page) values ($1, $2, $3, $4, $5, $6) ON CONFLICT (board, no) DO UPDATE SET last_modified = $3, page = $6"
	_, err := d.conn.Exec(stmt, tt.Board, tt.No, tt.LastModified, tt.LastArchived, tt.Replies, tt.Page)
	return err
}

func (d *dbClient) fetchThreadTask() ([]ThreadTask, error) {
	var tasks []ThreadTask
	now := time.Now()
	stmt := "SELECT no, board, last_modified, last_archived, replies, page FROM thread_backlog where last_modified > last_archived AND last_archived < $1 ORDER BY page ASC LIMIT 250"

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

func (d *dbClient) deleteThreadTask(id int) error {
	stmt := "DELETE FROM thread_backlog where id = $1"
	_, err := d.conn.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *dbClient) insertPost(no int, resto int, time int, name string, trip string, com string) error {
	stmt := "INSERT INTO post(no, resto, time, name, trip, com) values($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING;"
	_, err := d.conn.Exec(stmt, no, resto, time, name, trip, com)
	if err != nil {
		return err
	}
	return nil
}

func (d *dbClient) insertThread(no int, time int, name string, trip string, sub string, com string, replies int, images int) error {
	stmt := "INSERT INTO thread(no, time, name, trip, sub, com, replies, images) values($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING;"
	_, err := d.conn.Exec(stmt, no, time, name, trip, sub, com, replies, images)
	if err != nil {
		fmt.Println("xd")
		return err
	}
	return nil
}
