# Panopticon

Imageboard archiver


- init (spawns boardwatcher thread for each board)
- image processing worker: get sha256 hash and dimensions. insert sha256 and verify dimensions
	- fetch image task from db
	- check db for hash. download if doesnt exist
	- process image (get hash + dimensions)
	- write to db
- http request thread.  sends 1 request every n seconds (default 1).
	- if board watcher task,
		- fetchBoard()
			- send results to board worker thread.
	- if thread watcher task, 
		- fetchThread()
			- send thread to thread worker thread
	- if image download task, save file to specified location and send media to image processing worker
-  board watcher loop. 
	- run once every n seconds (default 10) and send task to request thread
-  board worker: keep track of which threads are archived/deleted/newly created. 
	- listen for catalog details from http request thread
	- upsert page number/replies/date modified in thread_job
- thread watcher
	- continuosly query for thread tasks from backlog where
		- last_archived is NULL or
		- last_modified > last_archived
		- last_archived <= now -10 seconds
		- and sort by page desc
	- send tasks to http request thread
-  thread worker(s):
	- listen for threads from http thread
	- deduplicate posts from shared set of hashed id + board
		- insert only posts not in set
	- if thread missing 404: delete thread task from db + delete thread from set
	- if valid thread: update thread backlog in db
	- if new images, add download task to db