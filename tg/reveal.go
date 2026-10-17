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
	LostUserSaver
	LastClosedQuestion(context.Context) (store.Question, error)
	Accounts(ctx context.Context, userIDs []int) (map[int]int, error)
}

type RevealHandler struct {
	Store RevealStore
}

func (h *RevealHandler) HandleResponse(context.Context, Message) error {
	return nil
}

func (h *RevealHandler) Name() string {
	return "reveal"
}

func (h *RevealHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/sendanswer"
}

func (h *RevealHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	q, err := h.Store.LastClosedQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("last closed question: %w", err)
	}

	answers, err := h.Store.Answers(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("answers: %w", err)
	}

	userIDs := make([]int, len(answers))
	for i, a := range answers {
		userIDs[i] = a.UserID
	}

	accs, err := h.Store.Accounts(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	for _, a := range answers {
		acc := accs[a.UserID]
		msgs = append(
			msgs,
			RevealMessage(
				a.UserID,
				q.Q,
				q.Opts[q.A],
				a.IsRight,
				acc,
			),
		)

		if acc <= 0 {
			err := h.Store.SaveLostUser(ctx, a.UserID)
			if err != nil {
				return nil, fmt.Errorf("save lost user %d: %w", a.UserID, err)
			}
		}
	}

	return msgs, nil
}

func RevealMessage(userID int, question string, answer string, isRight bool, acc int) (msg SendRichMessage) {
	rightMsg := "Вы ответили правильно!"
	if !isRight {
		rightMsg = "Вы ответили неправильно."
	}
	msg = SendRichMessage{
		ChatID: userID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,

					Text: fmt.Sprintf("%s Правильный ответ: %s!", question, answer),
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

	if acc <= 0 {
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: "Вы проиграли!",
		})
	}

	return
}
