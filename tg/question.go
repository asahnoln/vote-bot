package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type Prefix string

type Userser interface {
	Users(context.Context) (userIDs []store.User, err error)
}

type QuestionStore interface {
	Userser
	CurrentQuestion(context.Context) (store.Question, error)
	SaveLastMessageID(ctx context.Context, userID int, msgID int) error
}

type QuestionHandler struct {
	Store QuestionStore
}

func (h *QuestionHandler) Condition(upd Update) bool {
	return upd.CallbackQuery == nil
}

func (h *QuestionHandler) Name() string {
	return "question"
}

func (h *QuestionHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	if h.Store == nil {
		return nil, ErrStoreNil
	}

	uIDs, err := h.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	q, err := h.Store.CurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("question store: obj %#v: %w", q, err)
	}

	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: q.Q,
				},
				{
					Type:  InputRichBlockButtonsType,
					Align: AlignLeft,
				},
			},
		},
	}
	for _, o := range q.Opts {
		// TODO: TDD utility for button auto row placement
		msg.RichMessage.Blocks[1].Buttons = append(
			msg.RichMessage.Blocks[1].Buttons,
			RichMessageButton{
				Text:         o,
				CallbackData: fmt.Sprintf("%s%s:%s", OptCallbackDataPrefix, q.ID, o),
			},
		)
	}

	for _, u := range uIDs {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (h *QuestionHandler) HandleResponse(ctx context.Context, resp Message) error {
	h.Store.SaveLastMessageID(ctx, resp.Chat.ID, resp.MessageID)
	return nil
}
