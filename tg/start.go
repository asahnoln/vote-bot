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
	Store     StartStore
	TextStore TextByKeyer
	Token     string
}

func (h *StartHandler) HandleResponse(context.Context, any) error {
	return nil
}

func (h *StartHandler) Name() string {
	return "start"
}

func (h *StartHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/start "+h.Token
}

func (h *StartHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	ok, err := h.Store.FirstQuestionMissed(ctx)
	if err != nil {
		return nil, fmt.Errorf("first question missed: %w", err)
	}

	text := h.TextStore.TextByKey("welcomeToTheGame")
	if ok {
		text = h.TextStore.TextByKey("missedFirstQuestion")
	} else {
		err = h.Store.SaveUser(
			ctx,
			upd.Message.Chat.ID,
			fmt.Sprintf("%s %s @%s", upd.Message.Chat.FirstName, upd.Message.Chat.LastName, upd.Message.Chat.Username),
		)

		switch {
		case errors.Is(err, store.ErrUserAlreadySaved):
			text = h.TextStore.TextByKey("alreadyJoinedTheGame")
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
