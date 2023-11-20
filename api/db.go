package main

import (
	"database/sql"
	"fmt"
)

type dbClient struct {
	conn *sql.DB
}

func (d *dbClient) getBoards() ([]Board, error) {
	var boards []Board
	stmt := "SELECT * FROM boards where unlisted = false"
	rows, err := d.conn.Query(stmt)
	if err != nil {
		return boards, nil
	}
	defer rows.Close()
	for rows.Next() {
		var board Board
		if err := rows.Scan(&board.Board, &board.Unlisted); err != nil {
			return boards, err
		}
		boards = append(boards, board)
		// fmt.Printf("Retrieved task: %d Board: %s\n", task.No, task.Board)
	}
	fmt.Println(boards)
	if err = rows.Err(); err != nil {
		return boards, err
	}
	return boards, nil
}

func (d *dbClient) getThreads(after int64, before int64, count int, board string, sort string) ([]Thread, error) {
	var threads []Thread
	stmt := "SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread LIMIT 10"
	rows, err := d.conn.Query(stmt)
	if err != nil {
		return threads, nil
	}
	defer rows.Close()
	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.No, &thread.Time, &thread.Name, &thread.Trip, &thread.Sub,
			&thread.Replies, &thread.Images, &thread.Board, &thread.Tid); err != nil {
			return threads, err
		}
		threads = append(threads, thread)
		// fmt.Printf("Retrieved task: %d Board: %s\n", task.No, task.Board)
	}
	if err = rows.Err(); err != nil {
		return threads, err
	}
	return threads, nil
}
