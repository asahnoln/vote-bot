package tg

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/asahnoln/vote-bot/store"
)

type StartStore interface {
	FirstQuestionMissed(ctx context.Context) (bool, error)
	SaveUser(ctx context.Context, userID int, name string) error
}

type StartTextStore interface {
	TextWelcomeToTheGame() string
	TextMissedFirstQuestion() string
	TextAlreadyJoinedTheGame() string
}

type StartHandler struct {
	Store StartStore
	Text  StartTextStore
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

type StartHandlerText struct{}

func (StartHandlerText) TextWelcomeToTheGame() string {
	return "💸 Ваш стартовый капитал " + strconv.Itoa(store.AccountStartMoney) + "$\nСтавьте с умом, играйте смело — и пусть удача будет на вашей стороне!"
}

func (StartHandlerText) TextMissedFirstQuestion() string {
	return "Вы пропустили первый вопрос. Правила есть правила: в этот раз вы не участвуете."
}

func (StartHandlerText) TextAlreadyJoinedTheGame() string {
	return "Вы уже в игре!\nПовторный вход не требуется."
}

func (h *StartHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	ok, err := h.Store.FirstQuestionMissed(ctx)
	if err != nil {
		return nil, fmt.Errorf("first question missed: %w", err)
	}

	text := h.Text.TextWelcomeToTheGame()
	if ok {
		text = h.Text.TextMissedFirstQuestion()
	} else {
		err = h.Store.SaveUser(
			ctx,
			upd.Message.Chat.ID,
			fmt.Sprintf("%s %s @%s", upd.Message.Chat.FirstName, upd.Message.Chat.LastName, upd.Message.Chat.Username),
		)

		switch {
		case errors.Is(err, store.ErrUserAlreadySaved):
			text = h.Text.TextAlreadyJoinedTheGame()
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
