package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type RevealStore interface {
	LastClosedQuestion(context.Context) (store.Question, error)
	Answer(ctx context.Context, userID int, questionID string) (store.Answer, error)
	Account(context.Context) (int, error)
}

type RevealHandler struct {
	Store RevealStore
}

func (r RevealHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/sendanswer"
}

func (r RevealHandler) Handle(ctx context.Context, upd Update) (data URLer, err error) {
	q, _ := r.Store.LastClosedQuestion(ctx)
	// a, _ := r.Store.Answer(ctx, upd.Message.Chat.ID, q.ID)
	acc, _ := r.Store.Account(ctx)

	rightMsg := "Вы ответили правильно!"
	// if !a.IsRight {
	// 	rightMsg = "Вы ответили неправильно."
	// }

	msg := SendRichMessage{
		ChatID: upd.Message.Chat.ID,
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

	return msg, nil
}

func (r RevealHandler) Name() string {
	return "reveal"
}
