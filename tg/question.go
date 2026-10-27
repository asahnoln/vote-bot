package tg

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/asahnoln/vote-bot/store"
)

type Prefix string

type Userser interface {
	Users(context.Context) (userIDs []store.User, err error)
}
type CurrentQuestioner interface {
	CurrentQuestion(context.Context) (store.Question, error)
}

type QuestionStore interface {
	Adminser
	Userser
	CurrentQuestioner
	SaveLastMessageID(ctx context.Context, userID int, msgID int) error
	SaveCurrentQuestionSentAt(context.Context) error
}

type QuestionHandler struct {
	Store QuestionStore
}

func (h *QuestionHandler) Condition(upd Update) bool {
	return upd.Message.Text == SendCurrentQuestionCommand
}

func (h *QuestionHandler) Name() string {
	return "question"
}

func (h *QuestionHandler) Handle(ctx context.Context, upd Update) (msgs []Request, err error) {
	uIDs, err := h.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	q, err := h.Store.CurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("question store: obj %#v: %w", q, err)
	}

	err = h.Store.SaveCurrentQuestionSentAt(ctx)
	if err != nil {
		return nil, fmt.Errorf("question sent at store: obj %#v: %w", q, err)
	}

	msg := SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: q.Q,
				},
			},
		},
	}

	btns := SpreadButtons(q.Opts, q.ID, 4)
	for _, row := range btns {
		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type:    InputRichBlockButtonsType,
			Align:   AlignLeft,
			Buttons: row,
		})
	}

	for _, u := range uIDs {
		msg.ChatID = u.UserID
		msgs = append(msgs, msg)
	}

	msg = SendRichMessage{
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: "Сообщение только для админов",
				},
				{
					Type: InputRichBlockParagraphType,
					Text: "Выслан вопрос:",
				},
				{
					Type: InputRichBlockParagraphType,
					Text: q.Q,
				},
				{
					Type: InputRichBlockParagraphType,
					Text: "Варианты ответов:",
				},
				{
					Type: InputRichBlockParagraphType,
					Text: strings.Join(slices.Collect(maps.Keys(q.Opts)), ", "),
				},
			},
		},
	}
	admins, err := h.Store.Admins(ctx)
	if err != nil {
		return nil, fmt.Errorf("admins: %w", err)
	}

	for _, a := range admins {
		msg.ChatID = a.UserID
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (h *QuestionHandler) HandleResponse(ctx context.Context, resp any) error {
	// TODO: Check for error
	h.Store.SaveLastMessageID(ctx, resp.(Message).Chat.ID, resp.(Message).MessageID)
	return nil
}
