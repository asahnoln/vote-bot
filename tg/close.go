package tg

import (
	"context"
	"fmt"
)

type CloseStore interface {
	Userser
	CloseCurrentQuestion(ctx context.Context) error
}

type CloseHandler struct {
	Store CloseStore
}

func (h *CloseHandler) HandleResponse(context.Context, Message) error {
	return nil
}

func (h *CloseHandler) Name() string {
	return "close"
}

func (h *CloseHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/closequestion"
}

func (h *CloseHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	users, err := h.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	err = h.Store.CloseCurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("close current question: %w", err)
	}

	msg := EditMessageText{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: "Вопрос закрыт!",
				},
			},
		},
	}

	for _, u := range users {
		msg.ChatID = u.UserID
		msg.MessageID = u.LastMessageID
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
