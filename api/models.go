package main

type Board struct {
	Board    string
	Unlisted bool
}

type Op struct {
	No      int64
	Time    int64
	Name    string
	Trip    string
	Sub     string
	Replies int
	Images  int
	Board   string
	Tid     string
}

type Post struct {
	No    int64
	Resto int64
	Time  int64
	Name  string
	Trip  string
	Com   string
	Board string
	Tid   string
	Pid   string
}

type Thread struct {
	Op   Op
	Post []Post
}

type FileMapping struct {
	Filename   string
	Ext        string
	Identifier string
	No         int64
	Board      string
	FileID     int64
	Tid        string
	Pid        string
}

type File struct {
	Sha256 string
	Md5    string
	W      int
	H      int
	Fsize  int
	Mime   string
}
