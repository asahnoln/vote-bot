// package bot provides voting functions to play vote game
package bot

import "fmt"

type CurrentQuestioner interface {
	CurrentQuestion() (string, []string, error)
}

type QuestionOutputter interface {
	OutputQuestion(q string, opts []string) error
}

type Bot struct {
	QuestionStore  CurrentQuestioner
	QuestionOutput QuestionOutputter
}

func (b *Bot) CurrentQuestion() error {
	q, opts, err := b.QuestionStore.CurrentQuestion()
	if err != nil {
		return fmt.Errorf("question store: %w", err)
	}

	err = b.QuestionOutput.OutputQuestion(q, opts)
	if err != nil {
		return fmt.Errorf("question output: %w", err)
	}

	return nil
}
