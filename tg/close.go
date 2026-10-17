package tg

import (
	"context"
	"fmt"

	"github.com/asahnoln/vote-bot/store"
)

type CloseStore interface {
	Userser
	CurrentQuestioner
	Answerser
	CloseCurrentQuestion(ctx context.Context) error
}

type CloseHandler struct {
	Store CloseStore
}

func (h *CloseHandler) HandleResponse(context.Context, Message) error {
	return nil
}

func (h *CloseHandler) Name() string {
	return "close"
}

func (h *CloseHandler) Condition(upd Update) bool {
	return upd.Message.Text == "/closequestion"
}

func (h *CloseHandler) Handle(ctx context.Context, upd Update) (msgs []URLer, err error) {
	users, err := h.Store.Users(ctx)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	q, err := h.Store.CurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("current question: %w", err)
	}

	answers, err := h.Store.Answers(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("answers: %w", err)
	}

	answersMap := make(map[int]store.Answer)
	for _, a := range answers {
		answersMap[a.UserID] = a
	}

	err = h.Store.CloseCurrentQuestion(ctx)
	if err != nil {
		return nil, fmt.Errorf("close current question: %w", err)
	}

	for _, u := range users {
		msg := EditMessageText{
			RichMessage: InputRichMessage{
				Blocks: []InputRichBlock{
					{
						Type: InputRichBlockParagraphType,
						Text: q.Q,
					},
				},
			},
		}

		msg.ChatID = u.UserID
		msg.MessageID = u.LastMessageID

		if a, ok := answersMap[u.UserID]; ok {
			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
				Type: InputRichBlockParagraphType,
				Text: "Ваш ответ: " + a.A,
			})

			if a.Bet > 0 {
				msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
					Type: InputRichBlockParagraphType,
					Text: fmt.Sprintf("Ваша ставка: %d", a.Bet),
				})
			} else {
				msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
					Type: InputRichBlockParagraphType,
					Text: "Вы не успели сделать ставку...",
				})
				msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
					Type: InputRichBlockParagraphType,
					Text: "Ваш ответ не засчитан.",
				})
			}
		} else {
			msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
				Type: InputRichBlockParagraphType,
				Text: "Вы не успели ответить...",
			})
		}

		msg.RichMessage.Blocks = append(msg.RichMessage.Blocks, InputRichBlock{
			Type: InputRichBlockParagraphType,
			Text: "Вопрос закрыт!",
		})

		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func ClosingMessage(userID int, messageID int, question string, answer string, bet int) (msg EditMessageText) {
	msg = EditMessageText{
		ChatID:    userID,
		MessageID: messageID,
		RichMessage: InputRichMessage{
			Blocks: []InputRichBlock{
				{
					Type: InputRichBlockParagraphType,
					Text: question,
				},
				{
					Type: InputRichBlockParagraphType,
					Text: "Ваш ответ: " + answer,
				},
				{
					Type: InputRichBlockParagraphType,
					Text: fmt.Sprintf("Ваша ставка: %d", bet),
				},
				{
					Type: InputRichBlockParagraphType,
					Text: "Вопрос закрыт!",
				},
			},
		},
	}

	return
}
