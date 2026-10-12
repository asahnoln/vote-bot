package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type Prefix string

type QuestionStore interface {
	Users(context.Context) (userIDs []int, err error)
	CurrentQuestion(context.Context) (store.Question, error)
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

func (h *QuestionHandler) Handle(ctx context.Context, upd Update) (data []URLer, err error) {
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
		msg.ChatID = u
		data = append(data, msg)
	}
	return data, nil
}
