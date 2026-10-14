package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type RevealStore interface {
	Userser
	LastClosedQuestion(context.Context) (store.Question, error)
	Answer(ctx context.Context, userID int, questionID string) (store.Answer, error)
	Account(ctx context.Context, userID int) (int, error)
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
	users, err := r.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	q, err := r.Store.LastClosedQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("last closed question: %w", err)
	}

	for _, u := range users {
		a, err := r.Store.Answer(ctx, u.UserID, q.ID)
		if err != nil {
			return nil, fmt.Errorf("answer: %w", err)
		}

		acc, err := r.Store.Account(ctx, u.UserID)
		if err != nil {
			return nil, fmt.Errorf("account: %w", err)
		}

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
						Text: fmt.Sprintf("Ваш текущий счет - %d.", acc),
					},
				},
			},
		}

		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
