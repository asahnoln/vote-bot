package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type Answerser interface {
	Answers(ctx context.Context, questionID string) ([]store.Answer, error)
}

type RevealStore interface {
	Answerser
	LastClosedQuestion(context.Context) (store.Question, error)
	Accounts(ctx context.Context, userIDs []int) (map[int]int, error)
}

type RevealHandler struct {
	Store RevealStore
}

func (r *RevealHandler) HandleResponse(context.Context, Message) error {
	return nil
}

func (r *RevealHandler) Name() string {
	return "reveal"
}

func (r *RevealHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/sendanswer"
}

func (r *RevealHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	q, err := r.Store.LastClosedQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("last closed question: %w", err)
	}

	answers, err := r.Store.Answers(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("answers: %w", err)
	}

	userIDs := make([]int, len(answers))
	for i, a := range answers {
		userIDs[i] = a.UserID
	}

	accs, err := r.Store.Accounts(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	for _, a := range answers {
		rightMsg := "Вы ответили правильно!"
		if !a.IsRight {
			rightMsg = "Вы ответили неправильно."
		}

		msg := SendRichMessage{
			RichMessage: InputRichMessage{
				Blocks: []InputRichBlock{
					{
						Type: InputRichBlockParagraphType,
						Text: fmt.Sprintf("%s Правильный ответ: %s!", q.Q, q.Opts[q.A]),
					},
					{
						Type: InputRichBlockParagraphType,
						Text: rightMsg,
					},
					{
						Type: InputRichBlockParagraphType,
						Text: fmt.Sprintf("Ваш текущий счет - %d.", accs[a.UserID]),
					},
				},
			},
		}

		msg.ChatID = a.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
