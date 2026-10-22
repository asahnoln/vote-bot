package tg

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type StartStore interface {
	FirstQuestionMissed(ctx context.Context) (bool, error)
	SaveUser(ctx context.Context, userID int, name string) error
}

type StartHandler struct {
	Store StartStore
}

func (h *StartHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *StartHandler) Name() string {
	return "start"
}

func (h *StartHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/start"
}

const (
	TextWelcomeToTheGame     = "Добро пожаловать в игру! Ваш стартовый капитал - 1000$."
	TextMissedFirstQuestion  = "Вы пропустили первый обязательный вопрос и потому не можете быть допущены к игре."
	TextAlreadyJoinedTheGame = "Вы уже участвуете в игре!"
)

func (h *StartHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	ok, err := h.Store.FirstQuestionMissed(ctx)
	if err != nil {
		return nil, fmt.Errorf("first question missed: %w", err)
	}

	text := TextWelcomeToTheGame
	if ok {
		text = TextMissedFirstQuestion
	} else {
		err = h.Store.SaveUser(
			ctx,
			upd.Message.Chat.ID,
			fmt.Sprintf("%s %s @%s", upd.Message.Chat.FirstName, upd.Message.Chat.LastName, upd.Message.Chat.Username),
		)

		switch {
		case errors.Is(err, store.ErrUserAlreadySaved):
			text = TextAlreadyJoinedTheGame
		case err != nil:
			return nil, fmt.Errorf("save user: %w", err)
		}
	}

	msg := SendRichMessage{
		ChatID: upd.Message.Chat.ID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: text,
				},
			},
		},
	}
	msgs = append(msgs, msg)
	return msgs, nil
}
