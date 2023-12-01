package main

import (
	"database/sql"
	"fmt"
	"strings"
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

func (d *dbClient) getThreads(after int64, before int64, count int, board string, sort string, trip string, hasImage string) ([]Op, error) {
	var ops []Op
	stmt := "SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread WHERE board = $1 and time < $2 and time > $3"
	if hasImage == "true" {
		stmt = stmt + " and has_image = true"
	} else if hasImage == "false" {
		stmt = stmt + " and has_image = false"
	}
	stmt2 := fmt.Sprintf(" ORDER BY time %s LIMIT $4", sort)
	stmt = stmt + stmt2
	rows, err := d.conn.Query(stmt, board, before, after, count)
	if err != nil {
		fmt.Println(err)
		return ops, nil
	}
	defer rows.Close()
	for rows.Next() {
		var op Op
		if err := rows.Scan(&op.No, &op.Time, &op.Name, &op.Trip, &op.Sub,
			&op.Replies, &op.Images, &op.Board, &op.Tid); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	if err = rows.Err(); err != nil {
		return ops, err
	}
	return ops, nil
}

func (d *dbClient) getPosts(after int64, before int64, count int, board string, sort string, trip string, hasImage string, name string) ([]Post, error) {
	var posts []Post
	result := []any{}
	stmt := "SELECT no, resto, time, name, trip, com, board, tid, pid FROM pid WHERE "
	if board != "" {
		result = append(result, board)
		stmt += fmt.Sprintf("board = $%d and ", len(result))
	}
	if name != "" {
		result = append(result, name)
		stmt += fmt.Sprintf("name = $%d and ", len(result))
	}
	if trip != "" {
		result = append(result, trip)
		stmt += fmt.Sprintf("trip = $%d and ", len(result))
	}
	if after != 0 {
		result = append(result, after)
		stmt += fmt.Sprintf("time > $%d and ", len(result))
	}
	if before != 0 {
		result = append(result, before)
		stmt += fmt.Sprintf("time < $%d and ", len(result))
	}
	stmt = strings.TrimSuffix(stmt, "and ")
	if hasImage == "true" {
		stmt = stmt + " and has_image = true"
	} else if hasImage == "false" {
		stmt = stmt + " and has_image = false"
	}
	stmt2 := fmt.Sprintf(" ORDER BY time %s", sort)
	stmt = stmt + stmt2
	if count != 0 {
		result = append(result, count)
		stmt += fmt.Sprintf(" LIMIT $%d", len(result))
	}
	rows, err := d.conn.Query(stmt, result...)
	if err != nil {
		fmt.Println(err)
		return posts, nil
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.No, &post.Resto, &post.Time, &post.Name, &post.Trip, &post.Com,
			&post.Board, &post.Tid, &post.Pid); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil
}
func (d *dbClient) getThreadByID(tid string) (Op, error) {
	var op Op
	stmt := "SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread WHERE tid = $1 LIMIT 1"
	rows, err := d.conn.Query(stmt, tid)
	if err != nil {
		fmt.Println(err)
		return op, nil
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&op.No, &op.Time, &op.Name, &op.Trip, &op.Sub,
			&op.Replies, &op.Images, &op.Board, &op.Tid); err != nil {
			return op, err
		}
	}
	if err = rows.Err(); err != nil {
		return op, err
	}
	return op, nil
}

func (d *dbClient) getPostsByID(tid string, pid string) ([]Post, error) {
	var posts []Post
	var stmt string
	var rows *sql.Rows
	var err error

	if tid != "" && pid == "" {
		stmt = "SELECT no, resto, time, name, trip, com, board, tid, pid FROM post WHERE tid = $1"
		rows, err = d.conn.Query(stmt, tid)
	} else if tid == "" && pid != "" {
		stmt = "SELECT no, resto, time, name, trip, com, board, tid, pid FROM post WHERE pid = $2"
		rows, err = d.conn.Query(stmt, pid)
	} else if tid != "" && pid != "" {
		stmt = "SELECT no, resto, time, name, trip, com, board, tid, pid FROM post WHERE tid = $1 and pid = $2"
		rows, err = d.conn.Query(stmt, tid, pid)
	}

	if err != nil {
		fmt.Println(err)
		return posts, nil
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.No, &post.Resto, &post.Time, &post.Name, &post.Trip, &post.Com,
			&post.Board, &post.Tid, &post.Pid); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil
}

func (d *dbClient) getPostByResto(resto int64) ([]Post, error) {
	var posts []Post
	stmt := "SELECT no, resto, time, name, trip, com, board, tid, pid FROM post WHERE resto = $1"
	rows, err := d.conn.Query(stmt, resto)
	if err != nil {
		fmt.Println(err)
		return posts, nil
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.No, &post.Resto, &post.Time, &post.Name, &post.Trip, &post.Com,
			&post.Board, &post.Tid, &post.Pid); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil
}
func (d *dbClient) getFileMapping(id string, isThread bool) ([]FileMapping, error) {
	var fms []FileMapping
	var stmt string
	if isThread {
		stmt = "SELECT filename, ext, identifier, no, board, fileid from file_mapping where tid = $1"
	} else {
		stmt = "SELECT filename, ext, identifier, no, board, fileid from file_mapping where pid = $1"
	}
	rows, err := d.conn.Query(stmt, id)
	if err != nil {
		fmt.Println(err)
		return fms, err
	}
	defer rows.Close()
	for rows.Next() {
		var fm FileMapping
		if err := rows.Scan(&fm.Filename, &fm.Ext, &fm.Identifier, &fm.No, &fm.Board, &fm.FileID); err != nil {
			return fms, err
		}
		fms = append(fms, fm)
	}
	if err = rows.Err(); err != nil {
		return fms, err
	}
	return fms, nil
}

func (d *dbClient) getFileMeta(key int, sha256 string, md5 string, w int, h int, fsize int, mime string) (File, error) {
	var file File
	stmt := "SELECT sha256, md5, w, h, fsize, mime from file where "
	result := []any{}
	if key != -1 {
		result = append(result, key)
		stmt += fmt.Sprintf("id = $%d and ", len(result))
	}
	if sha256 != "" {
		result = append(result, sha256)
		stmt += fmt.Sprintf("sha256 = $%d and ", len(result))
	}
	if md5 != "" {
		result = append(result, md5)
		stmt += fmt.Sprintf("md5 = $%d and ", len(result))
	}
	if w != 0 {
		result = append(result, w)
		stmt += fmt.Sprintf("w = $%d and ", len(result))
	}
	if h != 0 {
		result = append(result, h)
		stmt += fmt.Sprintf("h = $%d and ", len(result))
	}
	if fsize != 0 {
		result = append(result, fsize)
		stmt += fmt.Sprintf("fsize = $%d and ", len(result))
	}
	if mime != "" {
		result = append(result, mime)
		stmt += fmt.Sprintf("mime = $%d and ", len(result))
	}
	stmt = strings.TrimSuffix(stmt, "and ")

	rows, err := d.conn.Query(stmt, result...)
	if err != nil {
		fmt.Println(err)
		return file, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&file.Sha256, &file.Md5, &file.W, &file.H, &file.Fsize, &file.Mime); err != nil {
			return file, err
		}
	}
	if err = rows.Err(); err != nil {
		return file, err
	}
	return file, nil
}

func (d *dbClient) getThreadByNo(no string) ([]Op, error) {
	var ops []Op
	stmt := "SELECT no, time, name, trip, sub, replies, images, board, tid FROM thread WHERE no = $1 LIMIT 1"
	rows, err := d.conn.Query(stmt, no)
	if err != nil {
		fmt.Println(err)
		return ops, nil
	}
	defer rows.Close()
	for rows.Next() {
		var op Op
		if err := rows.Scan(&op.No, &op.Time, &op.Name, &op.Trip, &op.Sub,
			&op.Replies, &op.Images, &op.Board, &op.Tid); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	if err = rows.Err(); err != nil {
		return ops, err
	}
	return ops, nil
}
