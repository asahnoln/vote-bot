package store

type Question struct {
	Q      string
	A      int
	Opts   []string
	Closed bool
	Order  int
}
