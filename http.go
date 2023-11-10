package main

import (
	"fmt"
	"strconv"
	"time"
)

func (a *Archiver) httpWorker() {
	fmt.Println("Starting http worker")
	for {
		task := <-a.httpWorkerChannel

		switch task.taskType {
		case BOARD:
			fmt.Println("HTTP: received board task: " + string(task.board))
			c, err := a.imageboard.fetchCatalog(task)
			fmt.Println("HTTP: Fetched catalog details")
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("HTTP: sending catalog content to boardWorkerChannel")
				a.boardWorkerChannel <- c
				fmt.Println("HTTP: sent catalog content to boardWorkerChannel")
			}
		case THREAD:
			fmt.Println("HTTP: received thread task" + strconv.Itoa(task.id))
			t, err := a.imageboard.fetchThread(task, &a.db)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("HTTP: sending thread content to theadWorkerChannel")
				a.threadWorkerChannel <- Thread{task.board, t}
				fmt.Println("HTTP: sent thread content to theadWorkerChannel")
			}
		case MEDIA:
			fmt.Println("HTTP: received media task " + task.filename)
			m, err := a.imageboard.fetchMedia(task, &a.db, a.lru)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("HTTP: sending media content to mediaWorkerChannel")
				a.mediaWorkerChannel <- m
				fmt.Println("HTTP: sent thread content to mediaWorkerChannel")
			}

		case BOARDMETA:
			m, err := a.imageboard.fetchBoards()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("HTTP: sending board content to boardMetaWorkerChannel")
				a.boardMetaWorkerChannel <- m
				fmt.Println("HTTP: sent board content to boardMetaWorkerChannel")
			}
		}
		fmt.Println("HTTP: sleeping 2 seconds")
		time.Sleep(2 * time.Second)
	}
}
