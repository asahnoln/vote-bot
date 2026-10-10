// package store provides common structs and errors for store implementations
package store

import "errors"

type Question struct {
	Q      string   `firestore:"q,omitempty"`
	A      int      `firestore:"a,omitempty"`
	Opts   []string `firestore:"opts,omitempty"`
	Closed bool     `firestore:"closed"`
	Order  int      `firestore:"order,omitempty"`
}

var ErrNoQuestionsFound = errors.New("no questions found")
