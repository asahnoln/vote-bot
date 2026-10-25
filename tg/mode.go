package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type ModeOptionalStore interface {
	Admins(ctx context.Context) ([]store.User, error)
	SetMode(ctx context.Context, mode store.Mode) error
}

type ModeHandler struct {
	Store ModeOptionalStore
	Mode  store.Mode
}

func (h *ModeHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *ModeHandler) Name() string {
	return "mode"
}

func (h *ModeHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/set"+string(h.Mode)+"mode"
}

func (h *ModeHandler) Handle(ctx context.Context, update Update) (msgs []Request, err error) {
	admins, err := h.Store.Admins(ctx)
	if err != nil {
		return nil, fmt.Errorf("admins: %w", err)
	}

	err = h.Store.SetMode(ctx, h.Mode)
	if err != nil {
		return nil, fmt.Errorf("set mode: %w", err)
	}

	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{Text: "Установлен режим вопросов: " + h.Mode.String()},
			},
		},
	}
	for _, u := range admins {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}
