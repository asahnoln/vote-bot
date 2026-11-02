package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type Adminser interface {
	Admins(ctx context.Context) ([]store.User, error)
}

type ModeOptionalStore interface {
	Adminser
	Userser
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
	err = h.Store.SetMode(ctx, h.Mode)
	if err != nil {
		return nil, fmt.Errorf("set mode: %w", err)
	}

	admins, err := h.Store.Admins(ctx)
	if err != nil {
		return nil, fmt.Errorf("admins: %w", err)
	}

	msgs = append(msgs, msgsWithTextForUsers(
		admins,
		"Установлен режим вопросов: "+h.Mode.String(),
	)...)

	if h.Mode == store.RealMode {
		users, err := h.Store.Users(ctx)
		if err != nil {
			return nil, fmt.Errorf("users: %w", err)
		}

		msgs = append(msgs, msgsWithTextForUsers(
			users,
			// TODO: Check text
			fmt.Sprintf("Ваш баланс сброшен до стартовых %d$. Началась реальная игра!", store.AccountStartMoney),
		)...)
	}

	return msgs, nil
}

func msgsWithTextForUsers(users []store.User, text string) (msgs []Request) {
	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: text,
				},
			},
		},
	}

	for _, u := range users {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	return msgs
}
