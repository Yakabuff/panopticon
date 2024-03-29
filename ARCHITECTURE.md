# Architecture:

- init
	- spawns watcher/worker goroutines and creates channels
- media watcher: get media jobs 
	- fetch image task from db
	- send results to http thread
- media worker: process media
	- queue thumbnails
	- update file with sha256 hash in db
	- delete media task from backlog
- http request thread.  sends 1 request every n seconds (default 1).
	- if board watcher task,
		- fetchBoard()
			- send results to board worker thread.
	- if thread watcher task, 
		- fetchThread()
			- send thread to thread worker thread
	- if image download task, save file to specified location and send media metadata to image processing worker
		- fetchMedia()
			- download and insert media metadata into db if not cached
-  board watcher loop. 
	- run once every n seconds (default 10) and send task to request thread
-  board worker: 
	- listen for catalog details from http request thread
	- insert threads to backlog to be fetched
- thread watcher
	- continuosly query for thread tasks from backlog where
		- last_archived is NULL or
		- last_modified > last_archived
		- last_archived <= now -10 seconds
		- and sort by page desc
	- send tasks to http request thread
-  thread worker(s):
	- listen for threads from http thread
	- deduplicate posts from shared redis set of base64(id + board + time)
		- insert only posts not in set
	- if thread missing 404: delete thread task from db + delete thread from set
	- if valid thread: update thread backlog in db
	- if image not in redis cache, add download task to db