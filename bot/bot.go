// package bot provides voting functions to play vote game
package bot

import (
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type CurrentQuestioner interface {
	CurrentQuestion() (store.Question, error)
}

type QuestionOutputter interface {
	OutputQuestion(q string, opts []string) error
}

type Question struct {
	Store  CurrentQuestioner
	Output QuestionOutputter
}

func (b *Question) Current() error {
	q, err := b.Store.CurrentQuestion()
	if err != nil {
		return fmt.Errorf("question store: %w", err)
	}

	err = b.Output.OutputQuestion(q.Q, q.Opts)
	if err != nil {
		return fmt.Errorf("question output: %w", err)
	}

	return nil
}

type AnswerSaver interface {
	SaveAnswer(uID string, a string, bet int) error
}

type Answer struct {
	Store AnswerSaver
}

func (b *Answer) Save(uID, a string, bet int) error {
	err := b.Store.SaveAnswer(uID, a, bet)
	if err != nil {
		return fmt.Errorf("save answer: %w", err)
	}

	return nil
}
