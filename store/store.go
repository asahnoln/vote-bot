// Package store provides common structs and errors for store implementations
package store

import (
	"errors"
	"time"
)

type Question struct {
	ID     string          `firestore:"id,omitempty"`
	Q      string          `firestore:"q,omitempty"`
	Opts   map[string]bool `firestore:"opts,omitempty"`
	Order  int             `firestore:"order"`
	SentAt time.Time       `firestore:"sent_at"`
}

type Answer struct {
	UserID  int    `firestore:"user_id"`
	IsRight bool   `firestore:"is_right"`
	Bet     int    `firestore:"bet"`
	A       string `firestore:"a"`
	AllIn   bool   `firestore:"all_in"`
}

type User struct {
	UserID         int    `firestore:"user_id"`
	Name           string `firestore:"name"`
	Acc            int
	LastMessageID  int     `firestore:"last_message_id"`
	Lost           bool    `firestore:"lost"`
	RightAnswerCnt int     `firestore:"right_answer_cnt"`
	AvgSpeed       float64 `firestore:"avg_speed"`
}

var (
	ErrNoOpenQuestionsLeft              = errors.New("no open questions left")
	ErrQuestionHasNoOpts                = errors.New("question has no opts")
	ErrQuestionHasNoRightAnswer         = errors.New("question has no right answer")
	ErrQuestionsOffsetMustNotBeNegative = errors.New("questions offset must not be negative")
	ErrUserAlreadySaved                 = errors.New("user already saved")
)
