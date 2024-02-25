package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/mapping"
)

// Query database for threads and posts concurrently
// Write date of last post/thread into file
type Indexer struct {
	index map[string]bleve.Index
	conn  *sql.DB
}

// TODO: fetch posts first, then index
// For each unique board, if not exist create and insert into index
// if exist, open index and insert
// Update map with index
func (i Indexer) indexDB() {
	for {
		num := i.indexOP()
		if num == 0 {
			time.Sleep(time.Second * 60)
		}
	}
	//go indexThreads()
}

// Index loop that fetches ops from DB and indexes into bleve index
// Fetch last_indexed timestamp from file. If not exist, use time 0
// Query threads > that timestamp
// Index threads in bleve
// Update file with new timestamp.  If not exist write file
func (i Indexer) indexOP() int {
	fileName := filepath.Join(os.Getenv("PATH_INDEX"), "LAST_OP_DATE")
	date, err := getLastDateFromFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	ops, err := i.getOPs(date)
	if err != nil {
		fmt.Println(err)
	}
	if len(ops) == 0 {
		return 0
	}
	i.insertOpsIndex(ops)
	// Update file with date of last op
	dateLast := ops[len(ops)-1].Time
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
	}
	// Clear file contents
	f.Truncate(0)
	f.Seek(0, 0)
	// Write date to file
	fmt.Println("writing date to file")
	b := []byte(strconv.FormatInt(dateLast, 10))
	_, err = f.Write(b)
	if err != nil {
		fmt.Println(b)
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return len(ops)
}

func (i Indexer) getOPs(after int64) ([]Op, error) {
	var ops []Op
	stmt := "SELECT sub, com, time, tid FROM thread WHERE "
	stmt += fmt.Sprintf("time > %d ", after)
	stmt += "ORDER BY time ASC, tid ASC "
	stmt += fmt.Sprintf("LIMIT %d", 1000)
	fmt.Println(stmt)
	rows, err := i.conn.Query(stmt)
	if err != nil {
		fmt.Println(err)
		return ops, nil
	}
	defer rows.Close()
	for rows.Next() {
		var op Op
		if err := rows.Scan(&op.Sub, &op.Com, &op.Time, &op.Tid); err != nil {
			return ops, err
		}
		ops = append(ops, op)
	}
	if err = rows.Err(); err != nil {
		return ops, err
	}
	// fmt.Println(ops)
	return ops, nil
}
func (i Indexer) getComments(after int64) ([]Post, error) {
	var posts []Post
	stmt := "SELECT com, time, pid FROM thread WHERE "
	stmt += fmt.Sprintf("time > %d ", after)
	stmt += "ORDER BY time ASC, tid ASC "
	stmt += fmt.Sprintf("LIMIT %d", 5)
	fmt.Println(stmt)
	rows, err := i.conn.Query(stmt)
	if err != nil {
		fmt.Println(err)
		return posts, nil
	}
	defer rows.Close()
	for rows.Next() {
		var po Post
		if err := rows.Scan(&po.Com, &po.Time, &po.Pid); err != nil {
			return posts, err
		}
		posts = append(posts, po)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	// fmt.Println(ops)
	return posts, nil
}

type Op struct {
	Sub  string
	Com  string
	Time int64
	Tid  string
}
type Post struct {
	Com  string
	Time int64
	Pid  string
}

func buildIndexMapping() (*mapping.IndexMappingImpl, error) {
	indexMapping := bleve.NewIndexMapping()

	err := indexMapping.AddCustomAnalyzer("asdf",
		map[string]interface{}{
			"type": `custom`,
			"char_filters": []interface{}{
				`html`,
			},
			"tokenizer": `whitespace`,
			"token_filters": []interface{}{
				`stemmer_en_snowball`,
			},
		})
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func getLastDateFromFile(file string) (int64, error) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// scan first line
	scanner.Scan()
	t := scanner.Text()
	err = scanner.Err()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	time, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return time, nil
}

// Get index handle from hashmap
// If exist, load index and insert Ops
// If not exist, create new index, index OPs and update hashmap
func (i Indexer) insertOpsIndex(ops []Op) error {
	index, ok := i.index["op"]
	var err error
	if !ok {
		fileName := filepath.Join(os.Getenv("PATH_INDEX"), "op")
		tidMapping := bleve.NewDocumentDisabledMapping()
		timeMapping := bleve.NewDocumentDisabledMapping()

		opMapping := bleve.NewDocumentMapping()

		opMapping.AddSubDocumentMapping("Tid", tidMapping)
		opMapping.AddSubDocumentMapping("Time", timeMapping)

		indexMapping := bleve.NewIndexMapping()

		indexMapping.AddDocumentMapping("op", opMapping)
		index, err = bleve.NewUsing(fileName, indexMapping, scorch.Name, scorch.Name, nil)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	batch := index.NewBatch()
	for j := 0; j < len(ops); j++ {
		fmt.Println("batch insert " + ops[j].Tid)
		fmt.Println(ops[j])
		batch.Index(ops[j].Tid, ops[j])
		if err != nil {
			return err
		}
	}
	fmt.Println(batch.String())
	err = index.Batch(batch)
	if err != nil {
		return err
	}
	return nil
}
