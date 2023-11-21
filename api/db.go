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
	stmt := fmt.Sprintf("SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread WHERE board = $1 and time < $2 and time > $3 ORDER BY time %s LIMIT $4", sort)
	rows, err := d.conn.Query(stmt, board, before, after, count)
	if err != nil {
		fmt.Println(err)
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
	}
	if err = rows.Err(); err != nil {
		return threads, err
	}
	return threads, nil
}

func (d *dbClient) getThreadByID(tid int64, board string) (Thread, error) {
	var thread Thread
	stmt := "SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread WHERE board = $1 and tid = $2"
	rows, err := d.conn.Query(stmt, tid)
	if err != nil {
		fmt.Println(err)
		return thread, nil
	}
	defer rows.Close()
	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.No, &thread.Time, &thread.Name, &thread.Trip, &thread.Sub,
			&thread.Replies, &thread.Images, &thread.Board, &thread.Tid); err != nil {
			return thread, err
		}
	}
	if err = rows.Err(); err != nil {
		return thread, err
	}
	return thread, nil
}

func (d *dbClient) getPostsByID(tid int64, board string) ([]Post, error) {
	var posts []Post
	stmt := "SELECT no, resto, time, name, trip, com, board FROM thread WHERE post = $1 and tid = $2"
	rows, err := d.conn.Query(stmt, tid)
	if err != nil {
		fmt.Println(err)
		return posts, nil
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.No, &post.Resto, &post.Time, &post.Name, &post.Trip, &post.Com,
			&post.Board); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil
}
