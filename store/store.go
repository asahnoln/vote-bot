// package store provides common structs and errors for store implementations
package store

import "errors"

type Question struct {
	ID     string   `firestore:"id,omitempty"`
	Q      string   `firestore:"q,omitempty"`
	A      int      `firestore:"a,omitempty"`
	Opts   []string `firestore:"opts,omitempty"`
	Closed bool     `firestore:"closed"`
	Order  int      `firestore:"order"`
}

var ErrNoQuestionsFound = errors.New("no questions found")
