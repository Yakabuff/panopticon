package main

type Board struct {
	Board    string
	Unlisted bool
}

type Thread struct {
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
