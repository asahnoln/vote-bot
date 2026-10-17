// Package store provides common structs and errors for store implementations
package store

import "errors"

type Question struct {
	ID    string   `firestore:"id,omitempty"`
	Q     string   `firestore:"q,omitempty"`
	A     int      `firestore:"a,omitempty"`
	Opts  []string `firestore:"opts,omitempty"`
	Order int      `firestore:"order"`
}

type Answer struct {
	UserID  int    `firestore:"user_id"`
	IsRight bool   `firestore:"is_right"`
	Bet     int    `firestore:"bet"`
	A       string `firestore:"a"`
}

type User struct {
	UserID        int    `firestore:"user_id"`
	Name          string `firestore:"name"`
	Acc           int
	LastMessageID int  `firestore:"last_message_id"`
	Lost          bool `firestore:"lost"`
}

var (
	ErrNoOpenQuestionsLeft              = errors.New("no open questions left")
	ErrQuestionHasNoOpts                = errors.New("question has no opts")
	ErrQuestionAnswerIndexIsOutOfBounds = errors.New("question answer index is out of bounds")
	ErrQuestionsOffsetMustNotBeNegative = errors.New("questions offset must not be negative")
	ErrUserAlreadySaved                 = errors.New("user already saved")
)
