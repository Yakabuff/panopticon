package main

type Board struct {
	Board    string
	Unlisted bool
}
type Boards struct {
	Boards []Board
}

type Op struct {
	No       int64
	Time     int64
	Name     string
	Trip     string
	Sub      string
	Com      string
	Replies  int
	Images   int
	Board    string
	Tid      string
	HasImage bool
}
type Ops struct {
	Ops     []Op
	Before  int64
	After   int64
	HasPrev bool
}

type Post struct {
	No       int64
	Resto    int64
	Time     int64
	Name     string
	Trip     string
	Com      string
	Board    string
	Tid      string
	Pid      string
	HasImage bool
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
