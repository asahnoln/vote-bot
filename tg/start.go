package tg

import (
	"context"
	"fmt"
)

type UserSaver interface {
	SaveUser(ctx context.Context, userID int) error
}

type StartHandler struct {
	Store UserSaver
}

func (h *StartHandler) Name() string {
	return "start"
}

func (h *StartHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/start"
}

func (h *StartHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	err = h.Store.SaveUser(ctx, upd.Message.Chat.ID)
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	msg := SendRichMessage{
		ChatID: upd.Message.Chat.ID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: "Добро пожаловать в игру!",
				},
			},
		},
	}
	msgs = append(msgs, msg)
	return msgs, nil
}
