package tg

import "context"

type CloseStore interface {
	Userser
	CloseCurrentQuestion(ctx context.Context) error
}

type CloseHandler struct {
	Store CloseStore
}

func (h *CloseHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	users, _ := h.Store.Users(ctx)
	h.Store.CloseCurrentQuestion(ctx)

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
